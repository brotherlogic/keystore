package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"

	pbd "github.com/brotherlogic/discovery/proto"
	pb "github.com/brotherlogic/keystore/proto"
	pbvs "github.com/brotherlogic/versionserver/proto"
	"google.golang.org/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

type integrationStatusGetter struct {
	setup *IntegrationSetup
}

func (i *integrationStatusGetter) getStatus(reg *pbd.RegistryEntry) *pb.StoreMeta {
	if i.setup.master.Registry == reg {
		meta, _ := i.setup.master.GetMeta(context.Background(), &pb.Empty{})
		return meta
	}

	for _, fol := range i.setup.followers {
		if fol.Registry == reg {
			meta, _ := fol.GetMeta(context.Background(), &pb.Empty{})
			return meta
		}
	}

	return nil
}

func (i *integrationStatusGetter) write(reg *pbd.RegistryEntry, req *pb.SaveRequest) error {
	if i.setup.master.Registry.Identifier == reg.Identifier {
		_, err := i.setup.master.Save(context.Background(), req)
		return err
	}

	for _, fol := range i.setup.followers {
		if fol.Registry.Identifier == reg.Identifier {
			_, err := fol.Save(context.Background(), req)
			return err
		}
	}

	return fmt.Errorf("Unable to locate %v", reg)

}

type integrationServerGetter struct {
	numFollowers int
}

func (i *integrationServerGetter) getServers() []*pbd.RegistryEntry {
	result := []*pbd.RegistryEntry{&pbd.RegistryEntry{Identifier: "iammaster"}}

	for in := 1; in <= i.numFollowers; in++ {
		result = append(result, &pbd.RegistryEntry{Identifier: fmt.Sprintf("iamfollower%v", in)})
	}

	return result
}

type integrationVersionWriter struct {
	version *pbvs.Version
}

func (i *integrationVersionWriter) write(ver *pbvs.Version) error {
	i.version = ver
	return nil
}
func (i *integrationVersionWriter) read() (*pbvs.Version, error) {
	return i.version, nil
}

//IntegrationSetup setup for itests
type IntegrationSetup struct {
	master    *KeyStore
	followers []*KeyStore
}

//InitIntegrationTest preps for tests
func InitIntegrationTest(numFollowers int) *IntegrationSetup {
	os.RemoveAll(".inttest")

	master := Init(".inttest/master/")
	master.SkipLog = true
	master.SkipIssue = true
	master.GoServer.Registry = &pbd.RegistryEntry{Identifier: "iammaster"}

	sg := &integrationServerGetter{numFollowers: numFollowers}
	master.serverGetter = sg

	setup := &IntegrationSetup{}
	statusGetter := &integrationStatusGetter{setup: setup}
	master.serverStatusGetter = statusGetter
	setup.master = master

	followers := []*KeyStore{}
	for i := 1; i <= numFollowers; i++ {
		follower := Init(fmt.Sprintf(".inttest/follower%v/", i))
		follower.SkipLog = true
		follower.SkipIssue = true
		follower.GoServer.Registry = &pbd.RegistryEntry{Identifier: fmt.Sprintf("iamfollower%v", i)}
		follower.serverStatusGetter = statusGetter
		follower.serverGetter = sg
		followers = append(followers, follower)
	}
	setup.followers = followers

	return setup
}

func TestBasicWrite(t *testing.T) {
	testSetup := InitIntegrationTest(2)

	data, _ := proto.Marshal(&pb.DeleteObject{Deletes: 5})
	_, err := testSetup.master.Save(context.Background(), &pb.SaveRequest{
		Key:   "test",
		Value: &google_protobuf.Any{Value: data}})

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	//Wait for the tasks to complete
	time.Sleep(time.Second)

	//Master should reflect write
	meta, err := testSetup.master.GetMeta(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("Read of meta failed: %v", err)
	}
	if meta.Version != 1 {
		t.Errorf("Meta has returned wrong: %v", meta)
	}

	//Follower should reflect write
	for i := 0; i < 2; i++ {
		_, err := testSetup.followers[i].GetMeta(context.Background(), &pb.Empty{})
		if err != nil {
			t.Fatalf("Read of follower meta failed: %v", err)
		}
	}

	// Follow up write
	data, _ = proto.Marshal(&pb.DeleteObject{Deletes: 20})
	_, err = testSetup.master.Save(context.Background(), &pb.SaveRequest{
		Key:   "test",
		Value: &google_protobuf.Any{Value: data}})

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	//Wait for the tasks to complete
	time.Sleep(time.Second * 5)

	//Master should reflect write
	meta, err = testSetup.master.GetMeta(context.Background(), &pb.Empty{})
	if err != nil {
		t.Fatalf("Read of meta failed: %v", err)
	}
	if meta.Version != 2 {
		t.Errorf("Meta has returned wrong: %v", meta)
	}
	if testSetup.master.saveRequests != 2 {
		t.Errorf("Wrong number of save requests")
	}

	//Follower should reflect write
	for i := 0; i < 2; i++ {
		_, err := testSetup.followers[i].GetMeta(context.Background(), &pb.Empty{})
		if err != nil {
			t.Fatalf("Read of follower meta failed: %v", err)
		}
	}

}
