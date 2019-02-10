package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"

	pb "github.com/brotherlogic/keystore/proto"
)

// Store is the basic store type
type Store struct {
	MemMutex     *sync.Mutex
	Mem          map[string][]byte
	Path         string
	Meta         *pb.StoreMeta
	Updates      []*pb.SaveRequest
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

	s := Store{Mem: make(map[string][]byte), Path: p, Meta: meta, MemMutex: &sync.Mutex{}}
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
	write, err := k.LocalSaveBytes(adjustKey(req.Key), req.Value.Value)
	if write > 0 {
		req.WriteVersion = k.Meta.Version
		k.Updates = append(k.Updates, req)
	}
	return err
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
	//Don't write if the proto matches
	data, err := k.LocalReadBytes(key)
	if err == nil && match(data, bytes) {
		return k.Meta.GetVersion(), nil
	}

	k.MemMutex.Lock()
	k.Mem[adjustKey(key)] = bytes
	k.MemMutex.Unlock()

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

func (k *Store) localSave(key string, m proto.Message) error {
	data, _ := proto.Marshal(m)
	if len(data) == 0 {
		return fmt.Errorf("Cannot save empty proto")
	}
	_, err := k.LocalSaveBytes(key, data)
	return err
}

//LocalReadBytes reads bytes
func (k *Store) LocalReadBytes(key string) ([]byte, error) {
	k.MemMutex.Lock()
	if val, ok := k.Mem[adjustKey(key)]; ok {
		k.MemMutex.Unlock()
		return val, nil
	}
	k.MemMutex.Unlock()

	data, err := ioutil.ReadFile(k.Path + adjustKey(key))

	if err != nil {
		k.MemMutex.Lock()
		k.Mem[key] = data
		k.MemMutex.Unlock()
	}

	return data, err
}
func (k *Store) localRead(key string, faker proto.Message) (proto.Message, error) {
	d, _ := k.LocalReadBytes(key)
	proto.Unmarshal(d, faker)
	return faker, nil
}
