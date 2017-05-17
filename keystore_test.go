package main

import (
	"context"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"

	pbd "github.com/brotherlogic/keystore/proto"
	pb "github.com/brotherlogic/keystore/testproto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

func TestBasicSave(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}

	s := InitTest(".testbasicsave", true)
	err := s.localSave("/test/path", tp)

	if err != nil {
		t.Errorf("Error in saving proto: %v", err)
	}
}

func TestBasicRead(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}

	s := InitTest(".testbasicread", true)
	s.localSave("/test/path", tp)

	m, _ := s.localRead("/test/path", &pb.TestProto{})
	tp2 := m.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Read is failing: %v", m)
	}
}

func TestAcrossServers(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	s := InitTest(".testacrossservers", true)
	s.localSave("/test/path", tp)

	s2 := InitTest(".testacrossservers", false)
	m, err := s2.localRead("/test/path", &pb.TestProto{})

	if err != nil {
		t.Fatalf("Read has returned error: %v", err)
	}

	tp2 := m.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Read after save is failing: %v", m)
	}
}

func TestReadViaRPC(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	bytes, _ := proto.Marshal(tp)
	s := InitTest(".testreadviarpc", true)
	s.Save(context.Background(), &pbd.SaveRequest{Key: "/test/path", Value: &google_protobuf.Any{Value: bytes}})
	resp, err := s.Read(context.Background(), &pbd.ReadRequest{Key: "/test/path"})
	if err != nil {
		t.Fatalf("Failed on RPC read: %v", err)
	}
	tp2 := &pb.TestProto{}
	err = proto.Unmarshal(resp.Value, tp2)

	if err != nil {
		t.Fatalf("Failed on proto unmarshal: %v", err)
	}

	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Failed on re-reading proto: %v", tp2)
	}
}

//InitTest prepares a test instance
func InitTest(path string, delete bool) *KeyStore {
	if delete {
		os.RemoveAll(path)
	}
	return Init(path)
}
