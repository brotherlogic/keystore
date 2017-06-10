package keystoreclient

import (
	"context"

	"github.com/golang/protobuf/proto"

	pbd "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

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
	_, err := c.linker.Save(context.Background(), &pbd.SaveRequest{Key: key, Value: &google_protobuf.Any{Value: bytes}})
	return err
}

// Load loads a proto
func (c *Keystoreclient) Load(key string, typ proto.Message) (proto.Message, error) {
	res, _ := c.linker.Read(context.Background(), &pbd.ReadRequest{Key: key})
	proto.Unmarshal(res.Value, typ)
	return typ, nil
}
