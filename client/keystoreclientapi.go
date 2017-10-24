package keystoreclient

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"google.golang.org/grpc"

	pb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

const (
	retries       = 5
	waitTimeBound = 10
)

type getIP func(servername string) (string, int)

//Prodlinker Production ready linker
type Prodlinker struct {
	getter getIP
}

//Save saves out the thingy
func (p *Prodlinker) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	err := errors.New("first pass fail")
	for i := 0; i < retries; i++ {
		ip, port := p.getter("keystore")
		if port > 0 {
			conn, err2 := grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithInsecure())
			err = err2

			if err == nil {
				defer conn.Close()

				store := pb.NewKeyStoreServiceClient(conn)
				return store.Save(ctx, req, grpc.FailFast(false))
			}
		}

		time.Sleep(time.Second * time.Duration(rand.Intn(waitTimeBound)))
	}

	return &pb.Empty{}, fmt.Errorf("Unable to save %v last error: %v", req.GetKey(), err)
}

//Read reads out the thingy
func (p *Prodlinker) Read(ctx context.Context, req *pb.ReadRequest) (*google_protobuf.Any, error) {
	for i := 0; i < retries; i++ {
		ip, port := p.getter("keystore")
		if port > 0 {
			conn, err := grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithInsecure())
			if err == nil {
				defer conn.Close()

				store := pb.NewKeyStoreServiceClient(conn)
				return store.Read(ctx, req, grpc.FailFast(false))

			}
		}
		time.Sleep(time.Second * time.Duration(rand.Intn(waitTimeBound)))
	}
	return nil, errors.New("Unable to read " + req.GetKey())
}

//GetClient gets a networked client
func GetClient(f getIP) *Keystoreclient {
	return &Keystoreclient{linker: &Prodlinker{getter: f}, retries: 5, backoffTime: time.Minute * 5}
}
