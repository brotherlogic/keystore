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

	pb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

// DoRegister does RPC registration
func (k KeyStore) DoRegister(server *grpc.Server) {
	pb.RegisterKeyStoreServiceServer(server, &k)
}

//Mote promotes or demotes this server
func (k KeyStore) Mote(master bool) error {
	t := time.Now()
	k.LogFunction("Mote", int32(time.Now().Sub(t).Nanoseconds()/1000000))
	return nil
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
	k.LogFunction("Save", int32(time.Now().Sub(t).Nanoseconds()/1000000))
	return &pb.Empty{}, nil
}

// Read reads a proto
func (k *KeyStore) Read(ctx context.Context, req *pb.ReadRequest) (*google_protobuf.Any, error) {
	t := time.Now()
	data, _ := k.LocalReadBytes(req.Key)
	k.LogFunction("Read", int32(time.Now().Sub(t).Nanoseconds()/1000000))
	return &google_protobuf.Any{Value: data}, nil
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
	var verbose = flag.Bool("verbose", false, "Show all output")
	flag.Parse()

	server := Init(*folder)

	//Turn off logging
	if !*verbose {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	server.PrepServer()
	server.RegisterServer("keystore", false)
	server.Serve()
}
