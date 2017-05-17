package main

import (
	"os"
	"testing"

	pb "github.com/brotherlogic/keystore/testproto"
)

func TestBasicSave(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}

	s := InitTest(".testbasicsave", true)
	err := s.Save("/test/path", tp)

	if err != nil {
		t.Errorf("Error in saving proto: %v", err)
	}
}

func TestBasicRead(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}

	s := InitTest(".testbasicread", true)
	s.Save("/test/path", tp)

	m, _ := s.Read("/test/path", &pb.TestProto{})
	tp2 := m.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Read is failing: %v", m)
	}
}

func TestAcrossServers(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	s := InitTest(".testacrossservers", true)
	s.Save("/test/path", tp)

	s2 := InitTest(".testacrossservers", false)
	m, err := s2.Read("/test/path", &pb.TestProto{})

	if err != nil {
		t.Fatalf("Read has returned error: %v", err)
	}

	tp2 := m.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Read after save is failing: %v", m)
	}
}

//InitTest prepares a test instance
func InitTest(path string, delete bool) *KeyStore {
	if delete {
		os.RemoveAll(path)
	}
	return Init(path)
}
