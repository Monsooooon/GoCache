package lru

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	maxBytes := 0
	lru := New(int64(maxBytes), nil)
	if lru == nil {
		t.Fatalf("lru is nil when create lru with maxBytes = %v", maxBytes)
	}

	maxBytes = -1
	lru = New(int64(maxBytes), nil)
	if lru == nil {
		t.Fatalf("lru is nil when create lru with maxBytes = %v", maxBytes)
	}

	maxBytes = 1
	lru = New(int64(maxBytes), nil)
	if lru == nil {
		t.Fatalf("lru is nil when create lru with maxBytes = %v", maxBytes)
	}
}

type String string

func (d String) Len() int64 {
	return int64(len(d))
}

func TestAddGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("1234"))

	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache key1=1234 failed")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3)) // k1, v1 should be removed here

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	evicted_keys := make([]string, 0)
	callback := func(key string, value Value) {
		evicted_keys = append(evicted_keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}
	if !reflect.DeepEqual(expect, evicted_keys) {
		t.Fatalf("Call OnEvicted failed, expect %s, got %s", expect, evicted_keys)
	}
}
