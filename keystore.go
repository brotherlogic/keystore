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

		data, err := k.masterGetter.Read(ctx, &pb.ReadRequest{Key: key.Key})
		if err != nil {
			return fmt.Errorf("Bad read %v -> %v", key, err)
		}
		_, err = k.store.LocalSaveBytes(key.Key, data.GetPayload().GetValue())
		if err != nil {
			return err
		}
	}

	k.store.Meta.Version = reg.GetVersion()
	return nil
}
