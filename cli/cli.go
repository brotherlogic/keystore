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
	pbd "github.com/brotherlogic/discovery/proto"
	pbk "github.com/brotherlogic/keystore/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/resolver"
)

func doDial(entry *pbd.RegistryEntry) (*grpc.ClientConn, error) {
	return grpc.Dial(entry.Ip+":"+strconv.Itoa(int(entry.Port)), grpc.WithInsecure())
}

func dialMaster(server string) (*grpc.ClientConn, error) {
	ip, port, err := utils.Resolve(server, "keystorecli")
	if err != nil {
		return nil, err
	}

	return doDial(&pbd.RegistryEntry{Ip: ip, Port: port})
}

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	client := keystoreclient.GetClient(dialMaster)
	if len(os.Args) == 1 {
		client.Save(context.Background(), "/testingkeytryagain2", &pb.Card{Text: "Testing222"})

		conn, err := grpc.Dial("discovery:///keystore", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Cannot dial master: %v", err)
		}
		defer conn.Close()

		registry := pbk.NewKeyStoreServiceClient(conn)
		res, err := registry.GetMeta(context.Background(), &pbk.Empty{})
		if err != nil {
			log.Fatalf("Error doing compare job: %v", err)
		}
		fmt.Printf("GOT %v", res)
	} else {
		conn, err := grpc.Dial("discovery:///keystore", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Cannot dial master: %v", err)
		}
		defer conn.Close()

		registry := pbk.NewKeyStoreServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		res, err := registry.Delete(ctx, &pbk.DeleteRequest{Key: os.Args[1]})
		//res, err := registry.Read(ctx, &pbk.ReadRequest{Key: os.Args[1]})
		if err != nil {
			log.Fatalf("Error on read: %v", err)
		}

		fmt.Printf("%v -> %v", res, err)
	}
}
