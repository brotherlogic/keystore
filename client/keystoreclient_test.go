package keystoreclient

import (
	"context"
	"testing"

	"github.com/brotherlogic/keystore/store"

	pbd "github.com/brotherlogic/keystore/proto"
	pb "github.com/brotherlogic/keystore/testproto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

func getTestClient(path string) *keystoreclient {
	return &keystoreclient{linker: &localLinker{s: store.KeyStore{Mem: make(map[string][]byte), Path: path}}}
}

type localLinker struct {
	s store.KeyStore
}

func (l *localLinker) Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error) {
	return l.s.Save(ctx, req)
}

func (l *localLinker) Read(ctx context.Context, req *pbd.ReadRequest) (*google_protobuf.Any, error) {
	return l.s.Read(ctx, req)
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
