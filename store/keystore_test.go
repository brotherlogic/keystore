package store

import (
	"log"
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

func TestIncrementOfMeta(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	s := InitTest(".testMetaIncrement", true)
	c1 := s.Meta.Version
	err := s.localSave("/test/path", tp)
	if err != nil {
		t.Fatalf("Error in doing save: %v", err)
	}
	c2 := s.Meta.Version
	if c1 == c2 {
		t.Errorf("Failed to update meta on save")
	}
}

func TestIncrementOfMetaWithDiff(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	s := InitTest(".testMetaIncrement", true)
	err := s.localSave("/test/path", tp)
	if err != nil {
		t.Fatalf("Error in doing save: %v", err)
	}
	c1 := s.Meta.Version
	tp2 := &pb.TestProto{Key: "Key", Value: "Value2"}
	err = s.localSave("/test/path", tp2)
	if err != nil {
		t.Fatalf("Error in saving second file: %v", err)
	}
	c2 := s.Meta.Version
	if c1 == c2 {
		t.Errorf("Meta version has not been incremented: %v", s.Meta)
	}
}

func TestMatchFailOnArrayDiff(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 5, 3}
	if match(a, b) {
		t.Errorf("Failure to match")
	}
}

func TestIncrementOfMetaWithNoDiff(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	s := InitTest(".testMetaIncrement", true)
	err := s.localSave("/test/path", tp)
	if err != nil {
		t.Fatalf("Error in doing save: %v", err)
	}
	c1 := s.Meta.Version
	tp2 := &pb.TestProto{Key: "Key", Value: "Value"}
	err = s.localSave("/test/path", tp2)
	if err != nil {
		t.Fatalf("Error in saving second file: %v", err)
	}
	c2 := s.Meta.Version
	if c1 != c2 {
		t.Errorf("Meta version has been incremented despite no diff: %v", s.Meta)
	}
}

func TestReadOfMetaOnReload(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	s := InitTest(".testMetaOnReload", true)
	err := s.localSave("/test/path", tp)
	if err != nil {
		t.Fatalf("Error in doing save: %v", err)
	}
	c1 := s.Meta.Version
	log.Printf("HERE %v", s.Meta)

	s2 := InitTest(".testMetaOnReload", false)
	c2 := s2.Meta.Version
	if c1 != c2 {
		t.Errorf("Meta has not been read on reload: %v", s2.Meta)
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
	ks := InitStore(path)
	return &ks
}
