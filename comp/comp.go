package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	pbdi "github.com/brotherlogic/discovery/proto"
	pb "github.com/brotherlogic/keystore/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func findServers() []*pbdi.RegistryEntry {
	conn, _ := grpc.Dial(utils.Discover, grpc.WithInsecure())
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rs, _ := registry.ListAllServices(ctx, &pbdi.ListRequest{}, grpc.FailFast(false))

	rets := make([]*pbdi.RegistryEntry, 0)
	for _, r := range rs.GetServices().GetServices() {
		if r.Name == "keystore" {
			rets = append(rets, r)
		}
	}

	return rets
}

func getKeys(s *pbdi.RegistryEntry) []string {
	conn, _ := grpc.Dial(s.GetIp()+":"+strconv.Itoa(int(s.GetPort())), grpc.WithInsecure())
	defer conn.Close()

	registry := pb.NewKeyStoreServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rs, err := registry.GetDirectory(ctx, &pb.GetDirectoryRequest{})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	return rs.GetKeys()
}

func read(s *pbdi.RegistryEntry, key string) int {
	conn, _ := grpc.Dial(s.GetIp()+":"+strconv.Itoa(int(s.GetPort())), grpc.WithInsecure())
	defer conn.Close()

	registry := pb.NewKeyStoreServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rs, err := registry.Read(ctx, &pb.ReadRequest{Key: key})

	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	val := 0
	for _, v := range rs.GetPayload().GetValue() {
		val += int(v)
	}
	return val
}

func main() {
	servers := findServers()
	var mainServer *pbdi.RegistryEntry
	for _, s := range servers {
		if s.GetMaster() {
			mainServer = s
		}
	}

	for _, key := range getKeys(mainServer) {
		fmt.Printf("Key: %v = %v\n", key, read(mainServer, key))
		for _, s := range servers {
			if !s.GetMaster() {
				fmt.Printf(" Key:%v = %v", key, read(s, key))
			}
		}
	}
}
