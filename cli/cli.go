package main

import (
	"log"

	pb "github.com/brotherlogic/cardserver/card"
	"github.com/brotherlogic/keystore/client"
)

func main() {
	client := keystoreclient.GetClient()
	err := client.Save("/testkey", &pb.Empty{})
	log.Printf("%v", err)
}
