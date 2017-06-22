package keystoreclient

import (
	"context"
	"errors"
	"log"
	"testing"

	pbd "github.com/brotherlogic/keystore/proto"
	pb "github.com/brotherlogic/keystore/testproto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

type countFail struct {
	l       *localLinker
	c       int
	retries int
}

//Save saves out a proto
func (c *countFail) Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error) {
	log.Printf("Count %v and %v", c.c, c.retries)
	if c.c < c.retries-1 {
		c.c++
		return nil, errors.New("Designed to fail")
	}

	return c.l.Save(ctx, req)
}

//Read reads a proto
func (c *countFail) Read(ctx context.Context, req *pbd.ReadRequest) (*google_protobuf.Any, error) {
	if c.c < c.retries-1 {
		c.c++
		return nil, errors.New("Designed to fail")
	}

	return c.l.Read(ctx, req)
}

func TestSaveAndLoad(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	client := GetTestClient(".testsaveandload")
	err := client.Save("/testkey", tp)

	if err != nil {
		t.Fatalf("Error in saving message %v", err)
	}

	tp2t, err := client.Read("/testkey", &pb.TestProto{})
	if err != nil || tp2t == nil {
		t.Fatalf("Error in loading message %v with %v", err, tp2t)
	}

	tp2 := tp2t.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Failure in returned proto: %v", tp2)
	}
}

func TestSaveFail(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	client := GetTestClient(".testsaveandload")
	linker := client.linker.(*localLinker)
	client.linker = &countFail{l: linker, c: 2}
	err := client.HardSave("/testkey", tp)

	if err != nil {
		t.Fatalf("Error in saving message %v", err)
	}

	tp2t, err := client.Read("/testkey", &pb.TestProto{})
	if err != nil || tp2t == nil {
		t.Fatalf("Error in loading message %v with %v", err, tp2t)
	}

	tp2 := tp2t.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Failure in returned proto: %v", tp2)
	}
}

func TestSaveFailHard(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	client := GetTestClient(".testsaveandload")
	linker := client.linker.(*localLinker)
	client.linker = &countFail{l: linker, retries: 10}
	err := client.HardSave("/testkey", tp)

	if err == nil {
		t.Fatalf("No error message on a hard save %v", err)
	}
}

func TestReadFail(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	client := GetTestClient(".testsaveandload")
	linker := client.linker.(*localLinker)
	client.linker = &countFail{l: linker, c: 0}
	err := client.HardSave("/testkey", tp)

	if err != nil {
		t.Fatalf("Error in saving message %v", err)
	}

	client.linker = &countFail{l: linker}
	tp2t, err := client.HardRead("/testkey", &pb.TestProto{})
	if err != nil || tp2t == nil {
		t.Fatalf("Error in loading message %v with %v", err, tp2t)
	}

	tp2 := tp2t.(*pb.TestProto)
	if tp2.Key != "Key" || tp2.Value != "Value" {
		t.Errorf("Failure in returned proto: %v", tp2)
	}
}

func TestReadFailHard(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	client := GetTestClient(".testsaveandload")
	linker := client.linker.(*localLinker)
	client.linker = &countFail{l: linker, c: 0}
	client.HardSave("/testkey", tp)

	client.linker = &countFail{l: linker, c: 1, retries: 100}
	_, err := client.HardRead("/testkey", &pb.TestProto{})
	if err == nil {
		t.Fatalf("No error message on a hard read %v", err)
	}
}
