package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"github.com/brotherlogic/keystore/client"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
	pbdi "github.com/brotherlogic/discovery/proto"
	pbk "github.com/brotherlogic/keystore/proto"
)

func findServer(name string) (string, int) {
	conn, _ := grpc.Dial(utils.Discover, grpc.WithInsecure())
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	rs, _ := registry.ListAllServices(context.Background(), &pbdi.Empty{}, grpc.FailFast(false))

	for _, r := range rs.Services {
		if r.Name == name {
			return r.Ip, int(r.Port)
		}
	}

	return "", -1
}

func main() {
	client := keystoreclient.GetClient(findServer)
	if len(os.Args) == 1 {
		err := client.Save("/testingkeytryagain", &pb.Card{Text: "Testing222"})
		log.Fatalf("Error: %v", err)

		host, port := findServer("keystore")
		if port > 0 {
			conn, err := grpc.Dial(host+":"+strconv.Itoa(port), grpc.WithInsecure())
			if err != nil {
				log.Fatalf("Cannot dial master: %v", err)
			}
			defer conn.Close()

			registry := pbk.NewKeyStoreServiceClient(conn)
			res, err := registry.GetMeta(context.Background(), &pbk.Empty{}, grpc.FailFast(false))
			if err != nil {
				log.Fatalf("Error doing compare job: %v", err)
			}
			fmt.Printf("GOT %v", res)
		}
	} else {
		t := time.Now()
		host, port := findServer("keystore")
		if port <= 0 {
			log.Fatalf("Error in locating keystore")
		}
		conn, err := grpc.Dial(host+":"+strconv.Itoa(port), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Cannot dial master: %v", err)
		}
		defer conn.Close()

		registry := pbk.NewKeyStoreServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		res, err := registry.Read(ctx, &pbk.ReadRequest{Key: os.Args[1]})
		if err != nil {
			log.Fatalf("Error on read: %v", err)
		}

		fmt.Printf("Read %v -> %v in %v (%v)", os.Args[1], res.GetPayload(), time.Now().Sub(t), res.GetReadTime())
	}
}
