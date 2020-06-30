package gocache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/gocache"

type HTTPPool struct {
	self     string // IP:#Port for this node
	basePath string // prefix, such as /gocache/
}

// NewHTTPPool creates an instance of HTTPPool with IP:#Port
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) logRequest(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handles all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.logRequest("Method %s, URL Path %s", r.Method, r.URL.Path)

	//  /<basepath>/<groupname>/<key> required
	// parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	// if len(parts) != 2 {
	// 	http.Error(w, "invalid request format, expect /<basepath>/<groupname>/<key>", http.StatusBadRequest)
	// 	return
	// }

	// get params from request url
	groupName := r.FormValue("groupname")
	key := r.FormValue("key")

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // Not good
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
