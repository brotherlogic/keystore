package main

import (
	"context"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"

	pbd "github.com/brotherlogic/discovery/proto"
	pb "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

type testServerGetter struct{}

func (serverGetter testServerGetter) getServers() []*pbd.RegistryEntry {
	return []*pbd.RegistryEntry{&pbd.RegistryEntry{Ip: "madeup", Port: 123}}
}

type testServerStatusGetter struct{}

func (serverStatusGetter testServerStatusGetter) getStatus(entry *pbd.RegistryEntry) *pb.StoreMeta {
	return &pb.StoreMeta{Version: 75}
}

func InitTest(p string) *KeyStore {
	os.RemoveAll(p)
	s := Init(p)
	s.SkipLog = true
	return s
}

func TestMoteSuccess(t *testing.T) {
	s := InitTest(".testMoteSuccess/")
	s.serverGetter = &testServerGetter{}
	s.serverStatusGetter = &testServerStatusGetter{}
	s.Store.Meta.Version = 100

	val := s.Mote(true)
	if val != nil {
		t.Errorf("Server has not accepted mote when it was way ahead of the pack")
	}
}

func TestMoteFail(t *testing.T) {
	s := InitTest(".testMoteFail/")
	s.serverGetter = &testServerGetter{}
	s.serverStatusGetter = &testServerStatusGetter{}
	s.Store.Meta.Version = 50

	val := s.Mote(true)
	if val == nil {
		t.Errorf("Server has not accepted mote when it was way behind")
	}
}

func TestGetDirectory(t *testing.T) {
	s := InitTest(".testGetDirectory/")

	emp, _ := proto.Marshal(&pb.Empty{})

	s.Save(context.Background(), &pb.SaveRequest{Key: "/madeup/key/one", Value: &google_protobuf.Any{Value: emp}})
	s.Save(context.Background(), &pb.SaveRequest{Key: "/madeup/key/two", Value: &google_protobuf.Any{Value: emp}})

	dir, err := s.GetDirectory(context.Background(), &pb.GetDirectoryRequest{})

	if err != nil {
		t.Fatalf("Error in getting directory: %v", err)
	}

	if len(dir.Keys) != 2 {
		t.Errorf("Error in retrieving directory listing: %v", dir)
	}

	found := false
	for _, k := range dir.Keys {
		if k == "madeup/key/one" {
			found = true
		}
	}

	if !found {
		t.Errorf("Unable to locate key: %v", dir.Keys)
	}
}
