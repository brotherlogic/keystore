package keystoreclient

import (
	"context"
	"errors"
	"strconv"
	"time"

	"google.golang.org/grpc"

	pb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

const ()

type getIP func(servername string) (string, int)

//Prodlinker Production ready linker
type Prodlinker struct {
	getter getIP
}

//Save saves out the thingy
func (p *Prodlinker) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	ip, port := p.getter("keystore")
	if port > 0 {
		conn, err := grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithInsecure())

		if err == nil {
			defer conn.Close()

			store := pb.NewKeyStoreServiceClient(conn)
			return store.Save(ctx, req)
		}
	}

	return &pb.Empty{}, errors.New("Unable to save " + ip)
}

//Read reads out the thingy
func (p *Prodlinker) Read(ctx context.Context, req *pb.ReadRequest) (*google_protobuf.Any, error) {
	ip, port := p.getter("keystore")
	conn, _ := grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithInsecure())
	defer conn.Close()

	store := pb.NewKeyStoreServiceClient(conn)
	return store.Read(ctx, req)
}

//GetClient gets a networked client
func GetClient(f getIP) *Keystoreclient {
	return &Keystoreclient{linker: &Prodlinker{getter: f}, retries: 5, backoffTime: time.Minute * 5}
}
