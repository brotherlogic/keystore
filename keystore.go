package main

import (
	"fmt"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/keystore/proto"
)

//Mote promotes or demotes this server
func (k *KeyStore) Mote(master bool) error {
	entries := k.serverGetter.getServers()
	for _, entry := range entries {
		meta := k.serverStatusGetter.getStatus(entry)

		if meta.GetVersion() > k.Meta.GetVersion() {
			return fmt.Errorf("We're too behind to be master")
		}
	}

	return nil
}

// GetDirectory gets a directory listing
func (k *KeyStore) GetDirectory(ctx context.Context, req *pb.GetDirectoryRequest) (*pb.GetDirectoryResponse, error) {
	return &pb.GetDirectoryResponse{Keys: k.Store.GetStored()}, nil
}
