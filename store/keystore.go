package store

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
)

// Store is the basic store type
type Store struct {
	Mem  map[string][]byte
	Path string
}

//LocalSaveBytes saves out a bunch of bytes
func (k *Store) LocalSaveBytes(key string, bytes []byte) error {
	k.Mem[key] = bytes

	fullpath := k.Path + key
	log.Printf("WRITING %v", fullpath)
	dir := fullpath[0:strings.LastIndex(fullpath, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
	ioutil.WriteFile(fullpath, bytes, 0644)

	return nil
}

func (k *Store) localSave(key string, m proto.Message) error {
	data, _ := proto.Marshal(m)
	return k.LocalSaveBytes(key, data)
}

//LocalReadBytes reads bytes
func (k *Store) LocalReadBytes(key string) ([]byte, error) {
	if _, ok := k.Mem[key]; ok {
		return k.Mem[key], nil
	}

	// Try to read from the fs
	log.Printf("Reading from file %v", k.Path+key)
	data, _ := ioutil.ReadFile(k.Path + key)
	k.Mem[key] = data

	return data, nil
}
func (k *Store) localRead(key string, faker proto.Message) (proto.Message, error) {
	d, _ := k.LocalReadBytes(key)
	proto.Unmarshal(d, faker)
	return faker, nil
}
