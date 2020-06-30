package gocache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	key := "Hello"
	expect := []byte(key)
	if val, _ := f.Get(key); !reflect.DeepEqual(expect, val) {
		t.Errorf("f.Get(%v) expected to return %v, but got %v", key, expect, val)
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestNewGroup(t *testing.T) {
	g := NewGroup("group1", 2048, GetterFunc(func(key string) ([]byte, error) {
		return nil, nil
	}))
	if expect, ok := groups["group1"]; !ok || expect != g {
		t.Errorf("cannot obtain group object after TestNewGroup")
	}
}

func TestGet(t *testing.T) {

	// record times of loading from underline db
	loadCounts := make(map[string]int, len(db))

	g := NewGroup("group", 2048, GetterFunc(
		func(key string) ([]byte, error) {
			if value, ok := db[key]; ok {
				log.Printf("[Slow DB] find value = %s for key = %s", key, value)
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(value), nil
			}
			return nil, fmt.Errorf("could not find key = %s in local db", key)
		}))

	for k, v := range db {
		if bv, err := g.Get(k); err != nil || bv.String() != v {
			t.Fatalf("failed to get value of key = %s, expect %s, got %s", k, v, bv.String())
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] != 1 {
			t.Fatalf("cache miss again after another Get of key %s", k)
		}
	}

	// Test unknow key
	unknow_key := "unknow"
	if _, err := g.Get(unknow_key); err == nil {
		t.Fatalf("should err when fetch key=%s, but got nil", unknow_key)
	}
}
