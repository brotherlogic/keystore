package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/keystore/store"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/keystore/proto"
)

//SStore Server type for keystore
type SStore struct {
	store.KeyStore
}

// DoRegister does RPC registration
func (k SStore) DoRegister(server *grpc.Server) {
	pb.RegisterKeyStoreServiceServer(server, &k.KeyStore)
}

//Init a keystore
func Init(p string) *SStore {
	ks := &SStore{store.KeyStore{GoServer: &goserver.GoServer{}, Mem: make(map[string][]byte), Path: p}}
	ks.Register = ks
	return ks
}

// ReportHealth alerts if we're not healthy
func (k SStore) ReportHealth() bool {
	return true
}

func main() {
	var folder = flag.String("folder", "", "The folder to use as a base")
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
