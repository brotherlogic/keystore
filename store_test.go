package main

import (
	"os"
	"testing"
)

func TestMatchFailOnArrayDiff(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 5, 3}
	if match(a, b) {
		t.Errorf("Failure to match")
	}
}

func TestMatchFailOnArrayDiffLength(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 2}
	if match(a, b) {
		t.Errorf("Failure to match")
	}
}

//InitStoreTest prepares a test instance
func InitStoreTest(path string, delete bool) *Store {
	if delete {
		os.RemoveAll(path)
	}
	ks := InitStore(path)
	return &ks
}

func TestLoad(t *testing.T) {
	s1 := InitStoreTest(".test_load", true)
	savestring := []byte("blah")
	_, err := s1.LocalSaveBytes("key", savestring)
	if err != nil {
		t.Fatalf("Failed save: %v", err)
	}

	s2 := InitStoreTest(".test_load", false)
	resp, _, err := s2.LocalReadBytes("key")
	if err != nil {
		t.Fatalf("Failed read: %v", err)
	}

	if string(resp) != string(savestring) {
		t.Errorf("Failed on read %v -> %v", string(savestring), string(resp))
	}
}

func TestDualSave(t *testing.T) {
	s1 := InitStoreTest(".test_load", true)
	savestring := []byte("blah")
	_, err := s1.LocalSaveBytes("key", savestring)
	if err != nil {
		t.Fatalf("Failed save: %v", err)
	}

	_, err = s1.LocalSaveBytes("key", savestring)
	if err != nil {
		t.Fatalf("Failed save: %v", err)
	}

	if s1.Meta.Version != 1 {
		t.Errorf("Dual save occured")
	}
}

func TestDeleteKey(t *testing.T) {
	s := InitStoreTest(".test_load", true)
	savestring := []byte("blah")
	s.Meta.DeletedKeys = append(s.Meta.DeletedKeys, "key")
	_, err := s.LocalSaveBytes("key", savestring)
	if err == nil {
		t.Errorf("Saved deleted key")
	}

	_, _, err = s.LocalReadBytes("key")
	if err == nil {
		t.Errorf("Loaded deleted key")
	}

	keys := s.GetStored()
	if len(keys) != 0 {
		t.Errorf("Deleted key is being listed")
	}

}

func TestListDeletedKey(t *testing.T) {
	s := InitStoreTest(".test_load", true)
	savestring := []byte("blah")
	_, err := s.LocalSaveBytes("/key", savestring)
	if err != nil {
		t.Errorf("Saved deleted key")
	}

	_, _, err = s.LocalReadBytes("/key")
	if err != nil {
		t.Errorf("Loaded deleted key")
	}

	s.Meta.DeletedKeys = append(s.Meta.DeletedKeys, "/key")

	keys := s.GetStored()
	if len(keys) != 0 {
		t.Errorf("Deleted key is being listed")
	}

}
