package keystoreclient

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc"

	pbdi "github.com/brotherlogic/discovery/proto"
	pb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

//Prodlinker Production ready linker
type Prodlinker struct{}

func getIP(servername string) (string, int) {
	conn, _ := grpc.Dial("192.168.86.64:50055", grpc.WithInsecure())
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	entry := pbdi.RegistryEntry{Name: servername}
	r, _ := registry.Discover(context.Background(), &entry)
	return r.Ip, int(r.Port)
}

//Save saves out the thingy
func (p *Prodlinker) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	ip, port := getIP("keystore")
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
	ip, port := getIP("keystore")
	conn, _ := grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithInsecure())
	defer conn.Close()

	store := pb.NewKeyStoreServiceClient(conn)
	return store.Read(ctx, req)
}

//GetClient gets a networked client
func GetClient() *Keystoreclient {
	return &Keystoreclient{linker: &Prodlinker{}}
}
