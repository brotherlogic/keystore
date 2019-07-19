package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"

	pb "github.com/brotherlogic/keystore/proto"
)

// Store is the basic store type
type Store struct {
	Path         string
	Meta         *pb.StoreMeta
	lastSnapshot int64
}

//InitStore builds out a store
func InitStore(p string) Store {
	meta := &pb.StoreMeta{}

	// Make the directory if we need to
	os.MkdirAll(p, 0777)

	data, err := ioutil.ReadFile(p + "/root.meta")
	if err == nil {
		proto.Unmarshal(data, meta)
	}

	s := Store{
		Path: p, Meta: meta,
	}
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

// GetStored gets all the local keys
func (k *Store) GetStored() []string {
	files := make([]string, 0)
	filepath.Walk(k.Path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() != "root.meta" {
			files = append(files, path[len(k.Path):])
		}
		return nil
	})
	return files
}

func adjustKey(key string) string {
	if !strings.HasPrefix(key, "/") && len(key) > 0 {
		return "/" + key
	}
	return key
}

//LocalSaveBytes saves out a bunch of bytes
func (k *Store) LocalSaveBytes(key string, bytes []byte) (int64, error) {
	for _, k := range k.Meta.DeletedKeys {
		if k == key {
			return 0, fmt.Errorf("Can't write to deleted key")
		}
	}

	//Don't write if the proto matches
	data, err := k.LocalReadBytes(key)
	if err == nil && match(data, bytes) {
		return k.Meta.GetVersion(), nil
	}

	fullpath := k.Path + adjustKey(key)
	dir := fullpath[0:strings.LastIndex(fullpath, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
	ioutil.WriteFile(fullpath, bytes, 0644)

	//Increment the version
	k.Meta.Version++
	k.saveMeta()

	return k.Meta.Version, nil
}

//LocalReadBytes reads bytes
func (k *Store) LocalReadBytes(key string) ([]byte, error) {
	for _, k := range k.Meta.DeletedKeys {
		if k == key {
			return []byte{}, fmt.Errorf("Cannot read deleted key")
		}
	}
	data, err := ioutil.ReadFile(k.Path + adjustKey(key))
	return data, err
}
