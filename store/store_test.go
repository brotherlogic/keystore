package store

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"

	pbk "github.com/brotherlogic/keystore/proto"
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

func BenchmarkBasicSave(b *testing.B) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}

	s := InitTest(".testbasicsave", true)

	for n := 0; n < b.N; n++ {
		err := s.localSave("/test/path", tp)

		if err != nil {
			b.Errorf("Error in saving proto: %v", err)
		}
		tp.Value = fmt.Sprintf("Value %v", n)
	}
}

func TestEmptySave(t *testing.T) {
	tp := &pb.TestProto{}

	s := InitTest(".testemptysave", true)
	err := s.localSave("/test/path", tp)

	if err == nil {
		t.Errorf("Empty proto was saved correctly")
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

func TestUpdatesWrittenVersion(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	s := InitTest(".testupdateswrittenversion", true)
	bytes, _ := proto.Marshal(tp)
	s.Save(&pbk.SaveRequest{Key: "/test/path1", Value: &google_protobuf.Any{Value: bytes}})
	s.Save(&pbk.SaveRequest{Key: "/test/path2", Value: &google_protobuf.Any{Value: bytes}})

	if len(s.Updates) != 2 {
		t.Fatalf("Updates have not been saved: %v", s.Updates)
	}

	if s.Updates[len(s.Updates)-1].WriteVersion != s.Meta.Version {
		t.Errorf("Mismatch in updates: (%v and %v) -> %v", s.Updates[len(s.Updates)-1].WriteVersion, s.Meta.Version, s)
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
