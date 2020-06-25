package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
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

// FDial fundamental dial
func fial(host string) (*grpc.ClientConn, error) {
	return grpc.Dial(host, grpc.WithInsecure())
}

func dialServer(ctx context.Context, servername string) (*grpc.ClientConn, error) {
	if servername == "discover" {
		return fial(fmt.Sprintf("%v:%v", utils.LocalIP, utils.RegistryPort))
	}

	conn, err := fial(utils.Discover)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	registry := pbd.NewDiscoveryServiceV2Client(conn)
	val, err := registry.Get(ctx, &pbd.GetRequest{Job: servername})
	if err != nil {
		return nil, err
	}

	// Pick a server at random
	servernum := rand.Intn(len(val.GetServices()))
	return fial(fmt.Sprintf("%v:%v", val.GetServices()[servernum].GetIp(), val.GetServices()[servernum].GetPort()))
}

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	client := keystoreclient.GetClient(dialServer)
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
