package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
	store               *store.Store
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
	transferError       string
	catchups            int64
	reads               int64
	fanouts             int64
	longestHardSync     time.Duration
	lastSuccessfulWrite time.Time
}

type prodVersionWriter struct {
	dial func(server string) (*grpc.ClientConn, error)
}

// This performs a fan out write
func (serverVersionWriter *prodVersionWriter) write(v *pbvs.Version) error {
	conn, err := serverVersionWriter.dial("versionserver")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbvs.NewVersionServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	_, err = client.SetVersion(ctx, &pbvs.SetVersionRequest{Set: v})
	return err
}

func (serverVersionWriter *prodVersionWriter) read() (*pbvs.Version, error) {
	conn, err := serverVersionWriter.dial("versionserver")
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

func (serverGetter *prodServerGetter) getServers() []*pbd.RegistryEntry {
	servers, err := utils.ResolveAll("keystore")
	if err != nil {
		return make([]*pbd.RegistryEntry, 0)
	}
	return servers
}

type prodServerStatusGetter struct {
	dial func(entry *pbd.RegistryEntry) (*grpc.ClientConn, error)
}

func (serverStatusGetter *prodServerStatusGetter) write(entry *pbd.RegistryEntry, req *pb.SaveRequest) error {
	conn, err := serverStatusGetter.dial(entry)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKeyStoreServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	_, err = client.Save(ctx, req)
	return err
}

func (serverStatusGetter *prodServerStatusGetter) getStatus(entry *pbd.RegistryEntry) *pb.StoreMeta {
	var result *pb.StoreMeta

	conn, err := serverStatusGetter.dial(entry)
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
func (k *KeyStore) DoRegister(server *grpc.Server) {
	pb.RegisterKeyStoreServiceServer(server, k)
}

// GetState gets the state of the server
func (k *KeyStore) GetState() []*pbgs.State {
	return []*pbgs.State{
		&pbgs.State{Key: "cores", Value: k.store.Meta.GetVersion()},
		&pbgs.State{Key: "states", Value: int64(k.state)},
		&pbgs.State{Key: "tfail", Value: k.transferFailCount},
		&pbgs.State{Key: "elapsed", Value: k.elapsed},
		&pbgs.State{Key: "corew", Value: k.coreWrites},
		&pbgs.State{Key: "fanw", Value: k.fanWrites},
		&pbgs.State{Key: "fans", Value: k.fanouts},
		&pbgs.State{Key: "terror", Text: k.transferError},
		&pbgs.State{Key: "catchups", Value: k.catchups},
		&pbgs.State{Key: "reads", Value: int64(k.reads)},
		&pbgs.State{Key: "longest_hard_sync", TimeDuration: k.longestHardSync.Nanoseconds()},
		&pbgs.State{Key: "last_write", TimeValue: k.lastSuccessfulWrite.Unix()},
	}
}

//Init a keystore
func Init(p string) *KeyStore {
	s := store.InitStore(p)
	ks := &KeyStore{GoServer: &goserver.GoServer{}, store: &s}
	ks.Register = ks
	ks.serverGetter = &prodServerGetter{}
	ks.serverStatusGetter = &prodServerStatusGetter{ks.DoDial}
	ks.serverVersionWriter = &prodVersionWriter{ks.DialMaster}
	ks.lastSuccessfulWrite = time.Now()
	return ks
}

func (k *KeyStore) fanoutWrite(req *pb.SaveRequest) {
	servers := k.serverGetter.getServers()
	t := time.Now()
	for _, server := range servers {
		k.fanouts++
		err := k.serverStatusGetter.write(server, req)
		if err != nil {
			k.transferError = fmt.Sprintf("%v", err)
			k.transferFailCount++
		}
	}
	k.elapsed = time.Now().Sub(t).Nanoseconds() / 1000000
}

func (k *KeyStore) storeTime(t time.Time) {
	duration := time.Now().Sub(t)
	if duration > k.longestHardSync {
		k.longestHardSync = duration
	}
}

//HardSync does a hard sync with an available keystore
func (k *KeyStore) HardSync() error {
	t := time.Now()
	defer k.storeTime(t)

	conn, err := k.DialMaster("keystore")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewKeyStoreServiceClient(conn)
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel1()
	meta, err := client.GetMeta(ctx1, &pb.Empty{})
	if err != nil {
		return err
	}

	// Pull the GetDirectory
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel2()
	dir, err := client.GetDirectory(ctx2, &pb.GetDirectoryRequest{})
	if err != nil {
		return err
	}

	// Process and Store
	for _, entry := range dir.GetKeys() {
		ctx3, cancel3 := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel3()
		t := time.Now()
		data, err := client.Read(ctx3, &pb.ReadRequest{Key: entry}, grpc.MaxCallRecvMsgSize(1024*1024*1024))
		if err != nil {
			return fmt.Errorf("Failure on %v: %v", entry, err)
		}
		if time.Now().Sub(t) > time.Second*30 {
			k.RaiseIssue(ctx3, "Long Read", fmt.Sprintf("It took %v to read %v", time.Now().Sub(t), entry), false)
		}
		k.store.LocalSaveBytes(entry, data.GetPayload().GetValue())
	}

	//Update the meta
	k.store.Meta.Version = meta.GetVersion()

	k.state = pb.State_SOFT_SYNC

	return nil
}

// Save a save request proto
func (k *KeyStore) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	if req.GetWriteVersion() == 0 {
		k.coreWrites++
	} else {
		k.fanWrites++
	}

	if time.Now().Sub(k.lastSuccessfulWrite) > time.Hour {
		k.RaiseIssue(ctx, "Keystore behind", fmt.Sprintf("%v has been behind for an hour", k.Registry.Identifier), false)
	}

	if req.GetWriteVersion() > k.store.Meta.GetVersion()+1 {
		k.catchups++
		k.state = pb.State_HARD_SYNC
		go k.HardSync()
		return &pb.Empty{}, errors.New("Unable to apply the write, going into HARD_SYNC mode")
	}

	k.lastSuccessfulWrite = time.Now()
	v, _ := k.store.LocalSaveBytes(req.Key, req.Value.Value)

	// Fanout the writes async
	if req.GetWriteVersion() == 0 {
		go k.serverVersionWriter.write(&pbvs.Version{Key: VersionKey, Value: v, Setter: k.Registry.Identifier + "-keystore"})
		req.WriteVersion = v
		go k.fanoutWrite(req)
	}

	return &pb.Empty{}, nil
}

// Read reads a proto
func (k *KeyStore) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	t := time.Now()
	data, _ := k.store.LocalReadBytes(req.Key)
	return &pb.ReadResponse{Payload: &google_protobuf.Any{Value: data}, ReadTime: time.Now().Sub(t).Nanoseconds() / 1000000}, nil
}

//GetMeta gets the metadata
func (k *KeyStore) GetMeta(ctx context.Context, req *pb.Empty) (*pb.StoreMeta, error) {
	return k.store.Meta, nil
}

// ReportHealth alerts if we're not healthy
func (k *KeyStore) ReportHealth() bool {
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
	server.RPCTracing = true
	server.RegisterServer("keystore", false)
	server.mote = *mote
	server.serverGetter = &prodServerGetter{server: server.Registry.GetIdentifier()}
	server.Serve()
}
