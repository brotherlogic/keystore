package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/keystore/store"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbd "github.com/brotherlogic/discovery/proto"
	pbgs "github.com/brotherlogic/goserver/proto"
	"github.com/brotherlogic/goserver/utils"
	pb "github.com/brotherlogic/keystore/proto"
	pbvs "github.com/brotherlogic/versionserver/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

const (
	//VersionKey the key to use in version store
	VersionKey = "github.com.brotherlogic.keystore"
)

type masterGetter interface {
	GetDirectory(ctx context.Context, in *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error)
	Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error)
}

type serverGetter interface {
	getServers() []*pbd.RegistryEntry
}

type serverStatusGetter interface {
	getStatus(*pbd.RegistryEntry) *pb.StoreMeta
	write(*pbd.RegistryEntry, *pb.SaveRequest) error
}

type serverVersionWriter interface {
	write(*pbvs.Version) error
	read() (*pbvs.Version, error)
}

// KeyStore the main server
type KeyStore struct {
	*goserver.GoServer
	*store.Store
	serverGetter        serverGetter
	serverStatusGetter  serverStatusGetter
	serverVersionWriter serverVersionWriter
	masterGetter        masterGetter
	state               pb.State
	mote                bool
	transferFailCount   int64
	elapsed             int64
	coreWrites          int64
	fanWrites           int64
}

type prodVersionWriter struct {
	resolver func(string) (string, int32, error)
}

// This performs a fan out write
func (serverVersionWriter prodVersionWriter) write(v *pbvs.Version) error {
	ip, port, err := serverVersionWriter.resolver("versionserver")
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbvs.NewVersionServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.SetVersion(ctx, &pbvs.SetVersionRequest{Set: v})
	return err
}

func (serverVersionWriter prodVersionWriter) read() (*pbvs.Version, error) {
	ip, port, err := serverVersionWriter.resolver("versionserver")
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbvs.NewVersionServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	val, err := client.GetVersion(ctx, &pbvs.GetVersionRequest{Key: VersionKey})
	return val.GetVersion(), err
}

type prodServerGetter struct {
	server string
}

func (serverGetter prodServerGetter) getServers() []*pbd.RegistryEntry {
	servers := make([]*pbd.RegistryEntry, 0)

	conn, err := grpc.Dial(utils.Discover, grpc.WithInsecure())
	if err == nil {
		defer conn.Close()
		client := pbd.NewDiscoveryServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		list, err := client.ListAllServices(ctx, &pbd.ListRequest{})
		if err == nil {
			for _, l := range list.GetServices().Services {
				if l.GetName() == "keystore" && l.GetIdentifier() != serverGetter.server {
					servers = append(servers, l)
				}
			}
		}
	}

	return servers
}

type prodServerStatusGetter struct{}

func (serverStatusGetter prodServerStatusGetter) write(entry *pbd.RegistryEntry, req *pb.SaveRequest) error {
	conn, err := grpc.Dial(entry.GetIp()+":"+strconv.Itoa(int(entry.GetPort())), grpc.WithInsecure())
	if err == nil {
		defer conn.Close()
		client := pb.NewKeyStoreServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err = client.Save(ctx, req)
		return err
	}
	return err
}

func (serverStatusGetter prodServerStatusGetter) getStatus(entry *pbd.RegistryEntry) *pb.StoreMeta {
	var result *pb.StoreMeta

	conn, err := grpc.Dial(entry.GetIp()+":"+strconv.Itoa(int(entry.GetPort())), grpc.WithInsecure())
	if err == nil {
		defer conn.Close()
		client := pb.NewKeyStoreServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		meta, err := client.GetMeta(ctx, &pb.Empty{})
		if err == nil {
			result = meta
		}
	}

	return result
}

// DoRegister does RPC registration
func (k KeyStore) DoRegister(server *grpc.Server) {
	pb.RegisterKeyStoreServiceServer(server, &k)
}

// GetState gets the state of the server
func (k *KeyStore) GetState() []*pbgs.State {
	return []*pbgs.State{
		&pbgs.State{Key: "core", Value: k.Store.Meta.GetVersion()},
		&pbgs.State{Key: "state", Value: int64(k.state)},
		&pbgs.State{Key: "tfail", Value: k.transferFailCount},
		&pbgs.State{Key: "elapsed", Value: k.elapsed},
		&pbgs.State{Key: "corew", Value: k.coreWrites},
		&pbgs.State{Key: "fanw", Value: k.fanWrites},
	}
}

//Init a keystore
func Init(p string) *KeyStore {
	s := store.InitStore(p)
	ks := &KeyStore{GoServer: &goserver.GoServer{}, Store: &s}
	ks.Register = ks
	ks.serverGetter = &prodServerGetter{}
	ks.serverStatusGetter = &prodServerStatusGetter{}
	ks.serverVersionWriter = &prodVersionWriter{resolver: utils.Resolve}
	return ks
}

func (k *KeyStore) fanoutWrite(req *pb.SaveRequest) {
	servers := k.serverGetter.getServers()
	k.Log(fmt.Sprintf("FANOUT: %v", servers))
	t := time.Now()
	for _, server := range servers {
		err := k.serverStatusGetter.write(server, req)
		if err != nil {
			k.transferFailCount++
		}
	}
	k.elapsed = time.Now().Sub(t).Nanoseconds()
}

//HardSync does a hard sync with an available keystore
func (k *KeyStore) HardSync() error {
	host, port := k.GetIP("keystore")
	conn, err := grpc.Dial(host+":"+strconv.Itoa(port), grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		return err
	}

	client := pb.NewKeyStoreServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	meta, err := client.GetMeta(ctx, &pb.Empty{})
	if err != nil {
		return err
	}

	// Pull the GetDirectory
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dir, err := client.GetDirectory(ctx, &pb.GetDirectoryRequest{})
	if err != nil {
		return err
	}

	// Process and Store
	for _, entry := range dir.GetKeys() {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		data, err := client.Read(ctx, &pb.ReadRequest{Key: entry}, grpc.MaxCallRecvMsgSize(1024*1024*1024))
		if err != nil {
			return err
		}
		k.LocalSaveBytes(entry, data.GetPayload().GetValue())
	}

	//Update the meta
	k.Meta.Version = meta.GetVersion()

	k.state = pb.State_SOFT_SYNC

	return nil
}

// Save a save request proto
func (k *KeyStore) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	t := time.Now()
	if req.GetWriteVersion() == 0 {
		k.coreWrites++
	} else {
		k.fanWrites++
	}

	if req.GetWriteVersion() > k.Store.Meta.GetVersion()+1 {
		k.state = pb.State_HARD_SYNC
		go k.HardSync()
		return &pb.Empty{}, errors.New("Unable to apply the write, going into HARD_SYNC mode")
	}

	v, _ := k.LocalSaveBytes(req.Key, req.Value.Value)

	k.Log(fmt.Sprintf("Prepping for fanout: %v", req.GetWriteVersion()))

	// Fanout the writes async
	if req.GetWriteVersion() == 0 {
		go k.serverVersionWriter.write(&pbvs.Version{Key: VersionKey, Value: v, Setter: k.Registry.Identifier + "-keystore"})
		req.WriteVersion = v
		k.Log(fmt.Sprintf("Doing a fanout write: %v", req.WriteVersion))
		go k.fanoutWrite(req)
	}

	k.LogFunction("Save", t)
	return &pb.Empty{}, nil
}

// Read reads a proto
func (k *KeyStore) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	t := time.Now()
	data, _ := k.LocalReadBytes(req.Key)
	k.LogFunction("Read", t)
	return &pb.ReadResponse{Payload: &google_protobuf.Any{Value: data}, ReadTime: time.Now().Sub(t).Nanoseconds() / 1000000}, nil
}

//GetMeta gets the metadata
func (k *KeyStore) GetMeta(ctx context.Context, req *pb.Empty) (*pb.StoreMeta, error) {
	return k.Meta, nil
}

// ReportHealth alerts if we're not healthy
func (k KeyStore) ReportHealth() bool {
	return true
}

func main() {
	var folder = flag.String("folder", "/media/keystore", "The folder to use as a base")
	var quiet = flag.Bool("quiet", false, "Show all output")
	var mote = flag.Bool("mote", true, "Allows us to mote")
	flag.Parse()

	server := Init(*folder)

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	server.PrepServer()
	server.RegisterServer("keystore", false)
	server.mote = *mote
	server.serverGetter = &prodServerGetter{server: server.Registry.GetIdentifier()}
	server.Serve()
}
