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
	"google.golang.org/protobuf/proto"
	"google.golang.org/grpc"

	pbd "github.com/brotherlogic/discovery/proto"
	pbgd "github.com/brotherlogic/godiscogs"
	pbk "github.com/brotherlogic/keystore/proto"
	ppb "github.com/brotherlogic/proxy/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"

	google_protobuf "github.com/golang/protobuf/ptypes/any"

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
		client.Save(context.Background(), "/github.com/brotherlogic/recordcollection/records/683829277", &pbrc.Record{Release: &pbgd.Release{Id: 14881954, InstanceId: 683829277}})

		conn, err := utils.LFDialServer(context.Background(), "keystore")
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		conn, err := utils.LFDialServer(ctx, "keystore")
		if err != nil {
			log.Fatalf("Cannot dial master: %v", err)
		}
		defer conn.Close()

		registry := pbk.NewKeyStoreServiceClient(conn)

		//res, err := registry.Delete(ctx, &pbk.DeleteRequest{Key: os.Args[1]})
		//res, err := registry.Read(ctx, &pbk.ReadRequest{Key: os.Args[1]})
		data := &ppb.GithubKey{Key: os.Args[2]}
		bytes, err := proto.Marshal(data)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		res, err := registry.Save(ctx, &pbk.SaveRequest{Key: os.Args[1], Value: &google_protobuf.Any{Value: bytes}})
		if err != nil {
			log.Fatalf("Error on read: %v", err)
		}

		fmt.Printf("%v -> %v", res, err)
	}
}
