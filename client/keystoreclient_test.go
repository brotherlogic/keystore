package keystoreclient

import (
	"testing"

	pb "github.com/brotherlogic/keystore/testproto"
)

func TestSaveAndLoad(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}
	client := GetTestClient(".testsaveandload")
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
