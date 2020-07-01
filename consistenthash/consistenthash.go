package consistenthash

import (
	"hash/crc32"
	"log"
	"sort"
	"strconv"
)

type HashFunc func([]byte) uint32

// Map contains all keys of hashed virtual machines
type Map struct {
	hash     HashFunc
	replicas int            // # of virual nodes (replica) for one real node
	keys     []int          // hashed value of virtual nodes. Sorted
	hashmap  map[int]string // map hashed value to real node's name
}

// NewMap creates a Map instance
// uses crc32.ChecksumIEEE if hash function is not provided
func NewMap(replicas int, hashfn HashFunc) *Map {
	m := &Map{
		replicas: replicas,
		hash:     hashfn,
		hashmap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash ring.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // generate #replicas virtual node
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashmap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Delete adds some keys to the hash ring.
func (m *Map) Delete(keys ...string) {
	var toDeleteKeys []int
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // generate #replicas virtual node
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			toDeleteKeys = append(toDeleteKeys, hash)
			delete(m.hashmap, hash)
		}
	}
	sort.Ints(toDeleteKeys)
	log.Printf("to delete: %v", toDeleteKeys)
	i, j := 0, 0
	for _, key := range m.keys {
		if j == len(toDeleteKeys) || key != toDeleteKeys[j] {
			m.keys[i] = key
			i++
		} else {
			j++
		}
	}
	m.keys = m.keys[:i]
	log.Printf("keys: %v", m.keys)
}

// Get fetch the name of next real node on the ring for a given key
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashmap[m.keys[idx%len(m.keys)]]
}
