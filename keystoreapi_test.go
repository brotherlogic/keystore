package main

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	pbd "github.com/brotherlogic/discovery/proto"
	pb "github.com/brotherlogic/keystore/proto"
	pbvs "github.com/brotherlogic/versionserver/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

type testFailGDMasterGetter struct{}

func (masterGetter testFailGDMasterGetter) GetDirectory(ctx context.Context, in *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error) {
	return nil, errors.New("Built to fail")
}

func (masterGetter testFailGDMasterGetter) Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error) {
	return nil, errors.New("Built to fail")
}

type testMasterGetter struct{}

func (masterGetter testMasterGetter) GetDirectory(ctx context.Context, in *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error) {
	return &pb.GetDirectoryResponse{Keys: []*pb.FileMeta{&pb.FileMeta{Key: "key1"}, &pb.FileMeta{Key: "key2"}}, Version: 123}, nil
}

func (masterGetter testMasterGetter) Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error) {
	td := &pb.Empty{}
	data, _ := proto.Marshal(td)
	return &pb.ReadResponse{Payload: &google_protobuf.Any{Value: data}}, nil
}

type testFailRMasterGetter struct{}

func (masterGetter testFailRMasterGetter) GetDirectory(ctx context.Context, in *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error) {
	return &pb.GetDirectoryResponse{Keys: []*pb.FileMeta{&pb.FileMeta{Key: "key1"}, &pb.FileMeta{Key: "key2"}}, Version: 123}, nil
}

func (masterGetter testFailRMasterGetter) Read(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error) {
	return nil, errors.New("Built to fail")
}

type testServerGetter struct{}

func (serverGetter testServerGetter) getServers() []*pbd.RegistryEntry {
	return []*pbd.RegistryEntry{&pbd.RegistryEntry{Ip: "madeup", Port: 123}}
}

type testServerStatusGetter struct{}

func (serverStatusGetter testServerStatusGetter) getStatus(entry *pbd.RegistryEntry) *pb.StoreMeta {
	return &pb.StoreMeta{Version: 75}
}

func (serverStatusGetter testServerStatusGetter) write(entry *pbd.RegistryEntry, sr *pb.SaveRequest) error {
	return nil
}

type testVersionWriter struct {
	written []*pbvs.Version
}

func (serverVersionWriter *testVersionWriter) write(version *pbvs.Version) error {
	serverVersionWriter.written = append(serverVersionWriter.written, version)
	return nil
}

func (serverVersionWriter *testVersionWriter) read() (*pbvs.Version, error) {
	return &pbvs.Version{Key: VersionKey, Value: 100}, nil
}

func InitTest(p string) *KeyStore {
	os.RemoveAll(p)
	s := Init(p)
	s.GoServer.Registry = &pbd.RegistryEntry{Identifier: "madeup"}
	s.SkipLog = true
	s.SkipIssue = true
	s.masterGetter = &testMasterGetter{}
	s.mote = true
	return s
}

func TestResync(t *testing.T) {
	s := InitTest(".testresync")
	err := s.resync()

	if err != nil {
		t.Errorf("Error on resync: %v", err)
	}
}

func TestResyncFailOnGD(t *testing.T) {
	s := InitTest(".testresync")
	s.masterGetter = &testFailGDMasterGetter{}
	err := s.resync()

	if err == nil {
		t.Errorf("No Error on resync: %v", err)
	}
}

func TestResyncFailOnR(t *testing.T) {
	s := InitTest(".testresync")
	s.masterGetter = &testFailRMasterGetter{}
	err := s.resync()

	if err == nil {
		t.Errorf("No Error on resync: %v", err)
	}
}

func TestWriteVersion(t *testing.T) {
	s := InitTest(".testVersionWriter")
	emp, _ := proto.Marshal(&pb.SaveRequest{Key: "testing"})
	s.Save(context.Background(), &pb.SaveRequest{Key: "madeup", Value: &google_protobuf.Any{Value: emp}})

	time.Sleep(time.Second)

}

func BenchmarkBasicSave(b *testing.B) {
	s := InitTest(".testbasicbench")

	emp, _ := proto.Marshal(&pb.Empty{})

	for n := 0; n < b.N; n++ {
		s.Save(context.Background(), &pb.SaveRequest{Key: "madeup", Value: &google_protobuf.Any{Value: emp}})
		time.Sleep(time.Millisecond * 100)
	}
}

func TestMoteSuccess(t *testing.T) {
	s := InitTest(".testMoteSuccess/")
	s.serverGetter = &testServerGetter{}
	s.serverStatusGetter = &testServerStatusGetter{}
	s.store.Meta.Version = 100

	val := s.Mote(context.Background(), true)
	if val != nil {
		t.Errorf("Server has not accepted mote when it was way ahead of the pack:%v", val)
	}
}

func TestMoteFail(t *testing.T) {
	s := InitTest(".testMoteFail/")
	s.serverGetter = &testServerGetter{}
	s.serverStatusGetter = &testServerStatusGetter{}
	s.store.Meta.Version = 50

	val := s.Mote(context.Background(), true)
	if val == nil {
		t.Errorf("Server has not accepted mote when it was way behind")
	}
}

func TestGetDirectory(t *testing.T) {
	s := InitTest(".testGetDirectory/")

	emp, _ := proto.Marshal(&pb.SaveRequest{Key: "testing"})

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
		if k.Key == "madeup/key/one" {
			found = true
		}
	}

	if !found {
		t.Errorf("Unable to locate key: %v", dir.Keys)
	}
}
