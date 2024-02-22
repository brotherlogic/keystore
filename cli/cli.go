package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brotherlogic/goserver/utils"
	keystoreclient "github.com/brotherlogic/keystore/client"
	"google.golang.org/grpc"

	pbrc "github.com/brotherlogic/recordcollection/proto"

	pbd "github.com/brotherlogic/discovery/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
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

func main() {
	client := keystoreclient.GetClient(dialServer)

	ctx, cancel := utils.ManualContext("testing", time.Minute)
	defer cancel()

	record := &pbrc.Record{}
	nr, _, err := client.Read(ctx, "/github.com/brotherlogic/recordcollection/records/365221500", record)
	if err != nil {
		log.Fatalf("Bad: %v", err)
	}

	log.Printf("%v", nr)
}
