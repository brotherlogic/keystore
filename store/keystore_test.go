package store

import (
	"os"
	"testing"

	pb "github.com/brotherlogic/keystore/testproto"
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

//InitTest prepares a test instance
func InitTest(path string, delete bool) *Store {
	if delete {
		os.RemoveAll(path)
	}
	ks := &Store{make(map[string][]byte), path}
	return ks
}
