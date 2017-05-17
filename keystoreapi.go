package main

import (
	"flag"
	"io/ioutil"
	"log"

	"google.golang.org/grpc"

	"github.com/brotherlogic/goserver"
	pb "github.com/brotherlogic/keystore/proto"
)

// DoRegister does RPC registration
func (k KeyStore) DoRegister(server *grpc.Server) {
	pb.RegisterKeyStoreServiceServer(server, &k)
}

//Init a keystore
func Init(p string) *KeyStore {
	ks := &KeyStore{&goserver.GoServer{}, make(map[string][]byte), p}
	ks.Register = ks
	return ks
}

// ReportHealth alerts if we're not healthy
func (k KeyStore) ReportHealth() bool {
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
