package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/brotherlogic/keystore/proto"
)

// Store is the basic store type
type Store struct {
	Path         string
	Meta         *pb.StoreMeta
	lastSnapshot int64
	fileMeta     map[string]*pb.FileMeta
	mainMutex    *sync.Mutex
}

// InitStore builds out a store
func InitStore(p string) Store {
	meta := &pb.StoreMeta{}

	// Make the directory if we need to
	os.MkdirAll(p, 0777)

	data, err := ioutil.ReadFile(p + "/root.meta")
	if err == nil {
		proto.Unmarshal(data, meta)
	}

	s := Store{
		Path:      p,
		Meta:      meta,
		fileMeta:  make(map[string]*pb.FileMeta),
		mainMutex: &sync.Mutex{},
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
func (k *Store) GetStored() []*pb.FileMeta {
	files := make([]*pb.FileMeta, 0)
	filepath.Walk(k.Path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() != "root.meta" {
			key := path[len(k.Path):]
			deleted := false
			for _, dKey := range k.Meta.DeletedKeys {
				if dKey == key {
					deleted = true
				}
			}
			if !deleted && !strings.HasSuffix(key, ".meta") {
				meta, err := k.readFileMeta(key, k.Path, nil)
				if err == nil {
					files = append(files, meta)
				}
			}
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

// LocalSaveBytes saves out a bunch of bytes
func (k *Store) LocalSaveBytes(key string, bytes []byte) (int64, error) {
	k.mainMutex.Lock()
	defer k.mainMutex.Unlock()
	for _, k := range k.Meta.DeletedKeys {
		if k == key {
			return 0, fmt.Errorf("Can't write to deleted key")
		}
	}

	//Don't write if the proto matches
	data, fileMeta, err := k.LocalReadBytes(key, true)
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
	fileMeta.Version++
	k.saveMeta()

	return k.Meta.Version, k.saveFileMeta(fullpath, fileMeta)
}

func (k *Store) saveFileMeta(dataPath string, fileMeta *pb.FileMeta) error {
	data, _ := proto.Marshal(fileMeta)
	return ioutil.WriteFile(dataPath+".meta", data, 0644)
}

func (k *Store) readFileMeta(key string, dataPath string, err error) (*pb.FileMeta, error) {
	if err != nil {
		return &pb.FileMeta{}, err
	}

	data, err := ioutil.ReadFile(dataPath + ".meta")

	fileMeta := &pb.FileMeta{Key: key}
	if err == nil {
		proto.Unmarshal(data, fileMeta)
	} else if os.IsNotExist(err) {
		// keep the empty metadata and move on
		err = nil
	}
	return fileMeta, err
}

// LocalReadBytes reads bytes
func (k *Store) LocalReadBytes(key string, locked bool) ([]byte, *pb.FileMeta, error) {
	if !locked {
		k.mainMutex.Lock()
		defer k.mainMutex.Unlock()
	}
	for _, k := range k.Meta.DeletedKeys {
		if k == key {
			return []byte{}, nil, status.Error(codes.OutOfRange, "Cannot read deleted key")
		}
	}
	data, err := ioutil.ReadFile(k.Path + adjustKey(key))
	fileMeta, err := k.readFileMeta(key, k.Path+adjustKey(key), err)
	return data, fileMeta, err
}
