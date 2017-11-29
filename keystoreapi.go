package main

import (
	"flag"
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
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

type serverGetter interface {
	getServers() []*pbd.RegistryEntry
}

type serverStatusGetter interface {
	getStatus(*pbd.RegistryEntry) *pb.StoreMeta
	write(*pbd.RegistryEntry, *pb.SaveRequest)
}

// KeyStore the main server
type KeyStore struct {
	*goserver.GoServer
	*store.Store
	serverGetter       serverGetter
	serverStatusGetter serverStatusGetter
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

		list, err := client.ListAllServices(ctx, &pbd.Empty{})
		if err == nil {
			for _, l := range list.Services {
				if l.GetName() == "keystore" && l.GetIdentifier() != serverGetter.server {
					log.Printf("IGNORE %v and %v", l.GetIdentifier(), serverGetter.server)
					servers = append(servers, l)
				}
			}
		}
	}

	return servers
}

type prodServerStatusGetter struct{}

func (serverStatusGetter prodServerStatusGetter) write(entry *pbd.RegistryEntry, req *pb.SaveRequest) {
	conn, err := grpc.Dial(entry.GetIp()+":"+strconv.Itoa(int(entry.GetPort())), grpc.WithInsecure())
	if err == nil {
		defer conn.Close()
		client := pb.NewKeyStoreServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		client.Save(ctx, req)
	}
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
	return []*pbgs.State{&pbgs.State{Key: "core", Value: k.Store.Meta.GetVersion()}}
}

//Init a keystore
func Init(p string) *KeyStore {
	s := store.InitStore(p)
	ks := &KeyStore{GoServer: &goserver.GoServer{}, Store: &s}
	ks.Register = ks
	ks.serverGetter = &prodServerGetter{}
	ks.serverStatusGetter = &prodServerStatusGetter{}
	return ks
}

func (k *KeyStore) fanoutWrite(req *pb.SaveRequest) {
	servers := k.serverGetter.getServers()
	for _, server := range servers {
		k.serverStatusGetter.write(server, req)
	}
}

// Save a save request proto
func (k *KeyStore) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	t := time.Now()
	v, _ := k.LocalSaveBytes(req.Key, req.Value.Value)

	// Fanout the writes async
	if req.GetWriteVersion() > 0 {
		req.WriteVersion = v
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
	var folder = flag.String("folder", "/media/disk1", "The folder to use as a base")
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	server := Init(*folder)

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	server.PrepServer()
	server.RegisterServer("keystore", false)
	server.serverGetter = &prodServerGetter{server: server.Registry.GetIdentifier()}
	server.Serve()
}
