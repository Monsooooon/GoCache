package gocache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	tt := []struct {
		name               string
		groupname          string
		key                string
		expectedStatusCode int
		expectedValue      string
		shouldValueExists  bool
	}{
		{"use valid key", "group", "Tom", http.StatusOK, db["Tom"], true},
		{"use invalid key", "group", "aom", http.StatusInternalServerError, "", false},
	}

	// create cache group and get the correct
	NewGroup("group", 2048, GetterFunc(
		func(key string) ([]byte, error) {
			if value, ok := db[key]; ok {
				log.Printf("[Slow DB] find value = %s for key = %s", key, value)

				return []byte(value), nil
			}
			return nil, fmt.Errorf("could not find key = %s in local db", key)
		}))

	// start HTTP Pool and mock request
	p := NewHTTPPool("localhost:8899")

	for _, tc := range tt {
		req, err := http.NewRequest("GET", p.basePath+"?groupname="+tc.groupname+"&key="+tc.key, nil)
		if err != nil {
			t.Fatalf("could not create request: %v", req)
		}
		log.Printf("Request URL Path = %s", req.URL)

		// create a mock ResponseWriter
		rec := httptest.NewRecorder()

		// do a mock http handling
		p.ServeHTTP(rec, req)

		// check response status
		res := rec.Result()
		defer res.Body.Close()
		if res.StatusCode != tc.expectedStatusCode {
			t.Fatalf("expect status %v, got %v", tc.expectedStatusCode, res.StatusCode)
		}

		// check response data
		res_b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("could not read response %v", err)
		}
		res_value := string(res_b)
		if tc.shouldValueExists && res_value != tc.expectedValue {
			t.Errorf("expect response value = %v, got %v", tc.expectedValue, res_value)
		}
		log.Printf("[TestServeHTTP] obtain value = %s for key = %s", res_value, tc.key)
	}

}
