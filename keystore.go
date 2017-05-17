package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/brotherlogic/goserver"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	pbd "github.com/brotherlogic/keystore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

// KeyStore the main server
type KeyStore struct {
	*goserver.GoServer
	mem  map[string][]byte
	path string
}

// Save a save request proto
func (k *KeyStore) Save(ctx context.Context, req *pbd.SaveRequest) (*pbd.Empty, error) {
	k.localSaveBytes(req.Key, req.Value.Value)
	return &pbd.Empty{}, nil
}

func (k *KeyStore) Read(ctx context.Context, req *pbd.ReadRequest) (*google_protobuf.Any, error) {
	data, _ := k.localReadBytes(req.Key)
	return &google_protobuf.Any{Value: data}, nil
}

func (k *KeyStore) localSaveBytes(key string, bytes []byte) error {
	k.mem[key] = bytes

	fullpath := k.path + key
	log.Printf("WRITING %v", fullpath)
	dir := fullpath[0:strings.LastIndex(fullpath, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
	ioutil.WriteFile(fullpath, bytes, 0644)

	return nil
}

func (k *KeyStore) localSave(key string, m proto.Message) error {
	data, _ := proto.Marshal(m)
	return k.localSaveBytes(key, data)
}

func (k *KeyStore) localReadBytes(key string) ([]byte, error) {
	if _, ok := k.mem[key]; ok {
		return k.mem[key], nil
	}

	// Try to read from the fs
	log.Printf("Reading from file %v", k.path+key)
	data, _ := ioutil.ReadFile(k.path + key)
	k.mem[key] = data

	return data, nil
}
func (k *KeyStore) localRead(key string, faker proto.Message) (proto.Message, error) {
	d, _ := k.localReadBytes(key)
	proto.Unmarshal(d, faker)
	return faker, nil
}
