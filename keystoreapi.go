package main

import (
	"flag"
	"io/ioutil"
	"log"

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

//Init a keystore
func Init(p string) *KeyStore {
	ks := &KeyStore{GoServer: &goserver.GoServer{}, Store: &store.Store{Mem: make(map[string][]byte), Path: p}}
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
	k.LocalSaveBytes(req.Key, req.Value.Value)
	return &pb.Empty{}, nil
}

// Read reads a proto
func (k *KeyStore) Read(ctx context.Context, req *pb.ReadRequest) (*google_protobuf.Any, error) {
	data, _ := k.LocalReadBytes(req.Key)
	return &google_protobuf.Any{Value: data}, nil
}

// ReportHealth alerts if we're not healthy
func (k KeyStore) ReportHealth() bool {
	return true
}

func main() {
	var folder = flag.String("folder", "/media/disk1/", "The folder to use as a base")
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
