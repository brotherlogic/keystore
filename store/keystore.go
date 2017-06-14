package store

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"

	pb "github.com/brotherlogic/keystore/proto"
)

// Store is the basic store type
type Store struct {
	Mem          map[string][]byte
	Path         string
	Meta         *pb.StoreMeta
	Updates      []*pb.SaveRequest
	lastSnapshot int64
}

//InitStore builds out a store
func InitStore(p string) Store {
	meta := &pb.StoreMeta{}
	data, err := ioutil.ReadFile(p + "/root.meta")
	log.Printf("READ: %v", data)
	if err == nil {
		proto.Unmarshal(data, meta)
	}

	s := Store{Mem: make(map[string][]byte), Path: p, Meta: meta}
	return s
}

func (k *Store) saveMeta() {
	data, _ := proto.Marshal(k.Meta)
	ioutil.WriteFile(k.Path+"/root.meta", data, 0644)
}

func match(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

//Save performs a local save
func (k *Store) Save(req *pb.SaveRequest) error {
	write, err := k.LocalSaveBytes(req.Key, req.Value.Value)
	if write {
		k.Updates = append(k.Updates, req)
	}
	return err
}

//LocalSaveBytes saves out a bunch of bytes
func (k *Store) LocalSaveBytes(key string, bytes []byte) (bool, error) {
	//Don't write if the proto matches
	data, err := k.LocalReadBytes(key)
	if err == nil && match(data, bytes) {
		return false, nil
	}

	k.Mem[key] = bytes

	fullpath := k.Path + key
	dir := fullpath[0:strings.LastIndex(fullpath, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
	ioutil.WriteFile(fullpath, bytes, 0644)

	//Increment the version
	log.Printf("HERE %v", k)
	log.Printf("ALSO %v", k.Meta)
	k.Meta.Version++
	k.saveMeta()

	return true, nil
}

func (k *Store) localSave(key string, m proto.Message) error {
	data, _ := proto.Marshal(m)
	_, err := k.LocalSaveBytes(key, data)
	return err
}

//LocalReadBytes reads bytes
func (k *Store) LocalReadBytes(key string) ([]byte, error) {
	if _, ok := k.Mem[key]; ok {
		return k.Mem[key], nil
	}

	data, _ := ioutil.ReadFile(k.Path + key)
	k.Mem[key] = data

	return data, nil
}
func (k *Store) localRead(key string, faker proto.Message) (proto.Message, error) {
	d, _ := k.LocalReadBytes(key)
	proto.Unmarshal(d, faker)
	return faker, nil
}
