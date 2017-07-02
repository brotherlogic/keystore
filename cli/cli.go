package main

import (
	"context"
	"log"
	"strconv"

	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
	pbdi "github.com/brotherlogic/discovery/proto"
	"github.com/brotherlogic/keystore/client"
	pbk "github.com/brotherlogic/keystore/proto"
)

func findServer(name string) (string, int) {
	conn, _ := grpc.Dial("192.168.86.64:50055", grpc.WithInsecure())
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	rs, _ := registry.ListAllServices(context.Background(), &pbdi.Empty{})

	for _, r := range rs.Services {
		if r.Name == name {
			log.Printf("%v -> %v", name, r)
			return r.Ip, int(r.Port)
		}
	}

	log.Printf("Could not find %v", name)
	return "", -1
}

func main() {
	client := keystoreclient.GetClient(findServer)
	err := client.Save("/testingkeytryagain", &pb.Card{Text: "Testing222"})
	log.Printf("Error: %v", err)

	host, port := findServer("keystore")
	if port > 0 {
		conn, err := grpc.Dial(host+":"+strconv.Itoa(port), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Cannot dial master: %v", err)
		}
		defer conn.Close()

		registry := pbk.NewKeyStoreServiceClient(conn)
		res, err := registry.GetMeta(context.Background(), &pbk.Empty{})
		if err != nil {
			log.Fatalf("Error doing compare job: %v", err)
		}
		log.Printf("GOT %v", res)
	}
}
