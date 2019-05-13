package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/keystore/proto"
)

// Shutdown the server
func (k *KeyStore) Shutdown(ctx context.Context) error {
	return nil
}

//Mote promotes or demotes this server
func (k *KeyStore) Mote(ctx context.Context, master bool) error {

	if !k.mote {
		return fmt.Errorf("Explicitly not moting, sorry")
	}

	entries := k.serverGetter.getServers()
	for _, entry := range entries {
		meta := k.serverStatusGetter.getStatus(entry)

		if meta.GetVersion() > k.store.Meta.GetVersion() {
			return fmt.Errorf("We're too behind to be master (versionstore says %v, we're %v)", meta.GetVersion(), k.store.Meta.GetVersion())
		}
	}

	//Check that we're up with version store
	vers, err := k.serverVersionWriter.read()
	if err != nil {
		return fmt.Errorf("Unable to determine where we are: %v", err)
	}
	if k.store.Meta.Version < vers.GetValue() {
		return fmt.Errorf("We're behind version store: %v and %v", k.store.Meta, vers)
	}

	if master {
		k.state = pb.State_MASTER
	}

	return nil
}

// GetDirectory gets a directory listing
func (k *KeyStore) GetDirectory(ctx context.Context, req *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error) {
	return &pb.GetDirectoryResponse{Keys: k.store.GetStored()}, nil
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
			return fmt.Errorf("Bad read %v -> %v", key, err)
		}
		_, err = k.store.LocalSaveBytes(key, data.GetPayload().GetValue())
		if err != nil {
			return err
		}
	}

	k.store.Meta.Version = reg.GetVersion()
	return nil
}
