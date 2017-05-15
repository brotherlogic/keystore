package main

import "testing"
import pb "github.com/brotherlogic/keystore/testproto"

func TestBasicSave(t *testing.T) {
	tp := &pb.TestProto{Key: "Key", Value: "Value"}

	s := Init()
	err := s.Save("/test/path", tp)

	if err != nil {
		t.Errorf("Error in saving proto: %v", err)
	}
}
