package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
)

// KeyStore the main server
type KeyStore struct {
	mem  map[string]proto.Message
	path string
}

//Init a keystore
func Init(p string) *KeyStore {
	ks := &KeyStore{}
	ks.mem = make(map[string]proto.Message)
	ks.path = p
	return ks
}

// Save a proto to the keystore
func (k *KeyStore) Save(key string, m proto.Message) error {
	k.mem[key] = m

	fullpath := k.path + key
	data, _ := proto.Marshal(m)
	log.Printf("WRITING %v", fullpath)
	dir := fullpath[0:strings.LastIndex(fullpath, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
	ioutil.WriteFile(fullpath, data, 0644)

	return nil
}

func (k *KeyStore) Read(key string, faker proto.Message) (proto.Message, error) {
	if _, ok := k.mem[key]; ok {
		return k.mem[key], nil
	}

	// Try to read from the fs
	log.Printf("Reading from file %v", k.path+key)
	data, _ := ioutil.ReadFile(k.path + key)
	proto.Unmarshal(data, faker)
	return faker, nil
}
