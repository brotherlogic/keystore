package keystoreclient

import (
	"context"
	"log"

	"github.com/golang/protobuf/proto"

	pbd "github.com/brotherlogic/keystore/proto"
	"github.com/brotherlogic/keystore/store"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

//GetTestClient gets a test client that saves to a local store
func GetTestClient(path string) *Keystoreclient {
	return &Keystoreclient{linker: &localLinker{s: store.Store{Mem: make(map[string][]byte), Path: path}}}
}

type localLinker struct {
	s store.Store
}

//Save saves out a proto
func (l *localLinker) Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error) {
	err := l.s.LocalSaveBytes(req.Key, req.Value.Value)
	return &pbd.Empty{}, err
}

//Read reads a proto
func (l *localLinker) Read(ctx context.Context, req *pbd.ReadRequest) (*google_protobuf.Any, error) {
	bytes, err := l.s.LocalReadBytes(req.Key)
	return &google_protobuf.Any{Value: bytes}, err
}

type link interface {
	Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error)
	Read(ctx context.Context, req *pbd.ReadRequest) (*google_protobuf.Any, error)
}

// Keystoreclient is the main client
type Keystoreclient struct {
	discovery string
	linker    link
}

// Save saves a proto
func (c *Keystoreclient) Save(key string, message proto.Message) error {
	bytes, _ := proto.Marshal(message)
	log.Printf("HERE %v", c.linker)
	_, err := c.linker.Save(context.Background(), &pbd.SaveRequest{Key: key, Value: &google_protobuf.Any{Value: bytes}})
	return err
}

// Load loads a proto
func (c *Keystoreclient) Load(key string, typ proto.Message) (proto.Message, error) {
	res, _ := c.linker.Read(context.Background(), &pbd.ReadRequest{Key: key})
	proto.Unmarshal(res.Value, typ)
	return typ, nil
}
