package keystoreclient

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/keystore/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	retries       = 5
	waitTimeBound = 10
)

type getIP func(ctx context.Context, servername string) (*grpc.ClientConn, error)

//Prodlinker Production ready linker
type Prodlinker struct {
	getter getIP
}

//Save saves out the thingy
func (p *Prodlinker) Save(ctx context.Context, req *pb.SaveRequest) (*pb.Empty, error) {
	conn, err := p.getter(ctx, "keystore")

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	store := pb.NewKeyStoreServiceClient(conn)
	return store.Save(ctx, req, grpc.FailFast(false))
}

//Read reads out the thingy
func (p *Prodlinker) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	conn, err := p.getter(ctx, "keystore")
	if err == nil {
		defer conn.Close()

		store := pb.NewKeyStoreServiceClient(conn)
		r, e := store.Read(ctx, req, grpc.FailFast(false), grpc.UseCompressor("gzip"))
		return r, e
	}
	code, ok := status.FromError(err)
	if ok {
		return nil, status.Errorf(code.Code(), "Unable to read %v last error: %v", req.GetKey(), err)
	}
	return nil, fmt.Errorf("Unable to read %v last error: %v", req.GetKey(), err)

}

//GetClient gets a networked client
func GetClient(f getIP) *Keystoreclient {
	return &Keystoreclient{linker: &Prodlinker{getter: f}, retries: 5, backoffTime: time.Minute * 5}
}
