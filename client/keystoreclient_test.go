package keystoreclient

import (
	"context"
	"testing"

	"github.com/brotherlogic/keystore/store"

	pbd "github.com/brotherlogic/keystore/proto"
	pb "github.com/brotherlogic/keystore/testproto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

func getTestClient(path string) *Keystoreclient {
	return &Keystoreclient{linker: &localLinker{s: store.Store{Mem: make(map[string][]byte), Path: path}}}
}

type localLinker struct {
	s store.Store
}

func (l *localLinker) Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error) {
	err := l.s.LocalSaveBytes(req.Key, req.Value.Value)
	return &pbd.Empty{}, err
}

func (l *localLinker) Read(ctx context.Context, req *pbd.ReadRequest) (*google_protobuf.Any, error) {
	bytes, err := l.s.LocalReadBytes(req.Key)
	return &google_protobuf.Any{Value: bytes}, err
}

func TestSaveAndLoad(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	client := getTestClient(".testsaveandload")
	err := client.Save("/testkey", tp)

	if err != nil {
		t.Fatalf("Error in saving message %v", err)
	}

	tp2t, err := client.Load("/testkey", &pb.TestProto{})
	if err != nil || tp2t == nil {
		t.Fatalf("Error in loading message %v with %v", err, tp2t)
	}

	tp2 := tp2t.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Failure in returned proto: %v", tp2)
	}
}
