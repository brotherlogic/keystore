package main

import (
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/keystore/store"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbgs "github.com/brotherlogic/goserver/proto"
	pb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

// DoRegister does RPC registration
func (k KeyStore) DoRegister(server *grpc.Server) {
	pb.RegisterKeyStoreServiceServer(server, &k)
}

//Mote promotes or demotes this server
func (k KeyStore) Mote(master bool) error {
	return nil
}

func (k KeyStore) GetState() []*pbgs.State {
	return []*pbgs.State{}
}

//Init a keystore
func Init(p string) *KeyStore {
	s := store.InitStore(p)
	ks := &KeyStore{GoServer: &goserver.GoServer{}, Store: &s}
	ks.Register = ks
	return ks
}

// KeyStore the main server
type KeyStore struct {
	*goserver.GoServer
	*store.Store
}

// Save a save request proto
func (k *KeyStore) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	t := time.Now()
	k.LocalSaveBytes(req.Key, req.Value.Value)
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
	server.Serve()
}
