package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	m := NewMap(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// generate virtual node at 2 12 22 4 14 24 6 16 26
	m.Add("2", "4", "6")

	tt := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"25": "6",
		"27": "2",
		"30": "2",
	}

	for k, v := range tt {
		if res := m.Get(k); v != res {
			t.Errorf("Get(%s) expects %s, got %s", k, v, res)
		}
	}

	// 2 4 6 8 12 14 16 18 22 24 26 28
	m.Add("8")
	tt["27"] = "8"
	for k, v := range tt {
		if res := m.Get(k); v != res {
			t.Errorf("After adding 8, Get(%s) expects %s, got %s", k, v, res)
		}
	}

	// 2 4 6 8 12 14 16 18 22 24 26 28
	// 2 6 8 12 16 18 22 26 28
	m.Delete("4")
	tt["23"] = "6"
	tt["13"] = "6"
	for k, v := range tt {
		if res := m.Get(k); v != res {
			t.Errorf("After deleting 4, Get(%s) expects %s, got %s", k, v, res)
		}
	}
}
