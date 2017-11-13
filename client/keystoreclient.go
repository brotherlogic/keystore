package keystoreclient

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang/protobuf/proto"

	pbd "github.com/brotherlogic/keystore/proto"
	"github.com/brotherlogic/keystore/store"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

//GetTestClient gets a test client that saves to a local store
func GetTestClient(path string) *Keystoreclient {
	return &Keystoreclient{linker: &localLinker{s: store.InitStore(path)}, retries: 5, backoffTime: time.Millisecond * 5}
}

type localLinker struct {
	s store.Store
}

//Save saves out a proto
func (l *localLinker) Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error) {
	_, err := l.s.LocalSaveBytes(req.Key, req.Value.Value)
	return &pbd.Empty{}, err
}

//Read reads a proto
func (l *localLinker) Read(ctx context.Context, req *pbd.ReadRequest) (*pbd.ReadResponse, error) {
	t := time.Now()
	bytes, err := l.s.LocalReadBytes(req.Key)
	return &pbd.ReadResponse{Payload: &google_protobuf.Any{Value: bytes}, ReadTime: time.Now().Sub(t).Nanoseconds() / 1000000}, err
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
}

// Save saves a proto
func (c *Keystoreclient) Save(key string, message proto.Message) error {
	bytes, _ := proto.Marshal(message)
	_, err := c.linker.Save(context.Background(), &pbd.SaveRequest{Key: key, Value: &google_protobuf.Any{Value: bytes}})
	return err
}

// Load loads a proto
func (c *Keystoreclient) Read(key string, typ proto.Message) (proto.Message, *pbd.ReadResponse, error) {
	res, err := c.linker.Read(context.Background(), &pbd.ReadRequest{Key: key})
	if err != nil {
		return nil, nil, err
	}
	proto.Unmarshal(res.GetPayload().GetValue(), typ)
	return typ, res, nil
}

// HardRead performs a read with retries
func (c *Keystoreclient) HardRead(key string, typ proto.Message) (proto.Message, *pbd.ReadResponse, error) {
	for i := 0; i < c.retries; i++ {
		v, val, err := c.Read(key, typ)
		if err == nil {
			return v, val, err
		}

		time.Sleep(c.backoffTime / time.Duration(c.retries))
	}

	return nil, nil, errors.New("Unable to perform hard read")
}

// HardSave performs a save with retries
func (c *Keystoreclient) HardSave(key string, message proto.Message) error {
	for i := 0; i < c.retries; i++ {
		log.Printf("TRY #%v", i)
		err := c.Save(key, message)
		log.Printf("SAVE ATTEMPT: %v", err)
		if err == nil {
			return err
		}

		time.Sleep(c.backoffTime / time.Duration(c.retries))
	}

	return errors.New("Unable to perform hard save")
}
