package main

import (
	"testing"

	pbd "github.com/brotherlogic/discovery/proto"
	pb "github.com/brotherlogic/keystore/proto"
)

type testServerGetter struct{}

func (serverGetter testServerGetter) getServers() []*pbd.RegistryEntry {
	return []*pbd.RegistryEntry{&pbd.RegistryEntry{Ip: "madeup", Port: 123}}
}

type testServerStatusGetter struct{}

func (serverStatusGetter testServerStatusGetter) getStatus(entry *pbd.RegistryEntry) *pb.StoreMeta {
	return &pb.StoreMeta{Version: 75}
}

func TestMoteSuccess(t *testing.T) {
	s := Init(".testMoteSuccess")
	s.serverGetter = &testServerGetter{}
	s.serverStatusGetter = &testServerStatusGetter{}
	s.Store.Meta.Version = 100

	val := s.Mote(true)
	if val != nil {
		t.Errorf("Server has not accepted mote when it was way ahead of the pack")
	}
}

func TestMoteFail(t *testing.T) {
	s := Init(".testMoteFail")
	s.serverGetter = &testServerGetter{}
	s.serverStatusGetter = &testServerStatusGetter{}
	s.Store.Meta.Version = 50

	val := s.Mote(true)
	if val == nil {
		t.Errorf("Server has not accepted mote when it was way behind")
	}
}
