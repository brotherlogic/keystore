package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/keystore/proto"
)

//Mote promotes or demotes this server
func (k *KeyStore) Mote(master bool) error {

	if !k.mote {
		return fmt.Errorf("Explicitly not moting, sorry")
	}

	entries := k.serverGetter.getServers()
	for _, entry := range entries {
		meta := k.serverStatusGetter.getStatus(entry)

		if meta.GetVersion() > k.Meta.GetVersion() {
			return fmt.Errorf("We're too behind to be master")
		}
	}

	//Check that we're up with version store
	vers, err := k.serverVersionWriter.read()
	if err != nil {
		return fmt.Errorf("Unable to determine where we are: %v", err)
	}
	if k.Store.Meta.Version < vers.GetValue() {
		return fmt.Errorf("We're behind version store: %v and %v", k.Store.Meta, vers)
	}

	if master {
		k.state = pb.State_MASTER
	}

	return nil
}

// GetDirectory gets a directory listing
func (k *KeyStore) GetDirectory(ctx context.Context, req *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error) {
	return &pb.GetDirectoryResponse{Keys: k.Store.GetStored()}, nil
}

func (k *KeyStore) resync() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	reg, err := k.masterGetter.GetDirectory(ctx, &pb.GetDirectoryRequest{})

	if err != nil {
		return err
	}

	for _, key := range reg.GetKeys() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		data, err := k.masterGetter.Read(ctx, &pb.ReadRequest{Key: key})
		if err != nil {
			return err
		}
		_, err = k.LocalSaveBytes(key, data.GetPayload().GetValue())
		if err != nil {
			return err
		}
	}

	k.Meta.Version = reg.GetVersion()
	return nil
}
