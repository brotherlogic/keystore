package keystoreclient

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbd "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

//GetTestClient gets a test client that saves to a local store
func GetTestClient(path string) *Keystoreclient {
	return &Keystoreclient{linker: &localLinker{store: make(map[string]*google_protobuf.Any)}, retries: 5, backoffTime: time.Millisecond * 5}
}

type localLinker struct {
	store map[string]*google_protobuf.Any
}

//Save saves out a proto
func (l *localLinker) Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error) {
	l.store[req.Key] = req.Value
	return &pbd.Empty{}, nil
}

//Read reads a proto
func (l *localLinker) Read(ctx context.Context, req *pbd.ReadRequest) (*pbd.ReadResponse, error) {
	if val, ok := l.store[req.Key]; ok {
		return &pbd.ReadResponse{Payload: val}, nil
	}
	return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Unable to locate %v", req.Key))
}

type link interface {
	Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error)
	Read(ctx context.Context, req *pbd.ReadRequest) (*pbd.ReadResponse, error)
}

// Keystoreclient is the main client
type Keystoreclient struct {
	discovery   string
	linker      link
	retries     int
	backoffTime time.Duration
	Fail        bool
}

// Save saves a proto
func (c *Keystoreclient) Save(ctx context.Context, key string, message proto.Message) error {
	bytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	_, err = c.linker.Save(ctx, &pbd.SaveRequest{Key: key, Value: &google_protobuf.Any{Value: bytes}})
	return err
}

// Load loads a proto
func (c *Keystoreclient) Read(ctx context.Context, key string, typ proto.Message) (proto.Message, *pbd.ReadResponse, error) {
	if c.Fail {
		return nil, nil, fmt.Errorf("Directed to fail")
	}
	res, err := c.linker.Read(ctx, &pbd.ReadRequest{Key: key})
	if err != nil {
		return nil, nil, err
	}
	proto.Unmarshal(res.GetPayload().GetValue(), typ)
	return typ, res, nil
}

// HardRead performs a read with retries
func (c *Keystoreclient) HardRead(ctx context.Context, key string, typ proto.Message) (proto.Message, *pbd.ReadResponse, error) {
	for i := 0; i < c.retries; i++ {
		v, val, err := c.Read(ctx, key, typ)
		if err == nil {
			return v, val, err
		}

		time.Sleep(c.backoffTime / time.Duration(c.retries))
	}

	return nil, nil, errors.New("Unable to perform hard read")
}

// HardSave performs a save with retries
func (c *Keystoreclient) HardSave(ctx context.Context, key string, message proto.Message) error {
	for i := 0; i < c.retries; i++ {
		err := c.Save(ctx, key, message)
		if err == nil {
			return err
		}

		time.Sleep(c.backoffTime / time.Duration(c.retries))
	}

	return errors.New("Unable to perform hard save")
}
