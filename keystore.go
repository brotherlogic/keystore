package main

import (
	"github.com/golang/protobuf/proto"
)

// KeyStore the main server
type KeyStore struct {
	mem map[string]proto.Message
}

//Init a keystore
func Init() *KeyStore {
	ks := &KeyStore{}
	ks.mem = make(map[string]proto.Message)
	return ks
}

// Save a proto to the keystore
func (k *KeyStore) Save(key string, m proto.Message) error {
	k.mem[key] = m
	return nil
}

func (k *KeyStore) Read(key string) (proto.Message, error) {
	return k.mem[key], nil
}
