package keystoreclient

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc"

	pb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

//Prodlinker Production ready linker
type Prodlinker struct {
	getIP func(servername string) (string, int)
}

//Save saves out the thingy
func (p *Prodlinker) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	ip, port := p.getIP("keystore")
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
	ip, port := p.getIP("keystore")
	conn, _ := grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithInsecure())
	defer conn.Close()

	store := pb.NewKeyStoreServiceClient(conn)
	return store.Read(ctx, req)
}

//GetClient gets a networked client
func GetClient() *Keystoreclient {
	return &Keystoreclient{linker: &Prodlinker{}}
}
