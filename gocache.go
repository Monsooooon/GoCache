package gocache

import (
	"fmt"
	"log"
	"sync"
)

/*

Receive key --> Check is key stored -----> Return bue
                		|  No         Yes
                		└-----> Check if should fetch bue from peers -----> interact with peer --> Return bue
                            |  No 										 Yes
                            └-----> Call callback to fetch bue from DB/File and add it into cache --> Return bue
*/

// Getter interface loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc type implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key) // CAUTION!!! NOT f.Get(key)
}

/*
	题外话：为什么设计 GetFunc 和它的 method Get，而不是只保留一个Getter接口供其他模块实现呢？
	思考：
		首先，当然是可以只保留 Gettter接口，对于各个获取数据的模块（dbGetter, fileGetter, ...)
		那么每个模块都实现 Gettter的Get函数就好了。

		但是假如我们设计了一个Get函数的实现。它仅仅是一个函数，例如：
		mygetter := func(key string) ([]byte, error) {
			...
		}
		我们能不能让一个Getter接口的变量等于它呢？即

		var g Getter = mygetter  // WRONG!!!!!!

		答案是不行！因为仅仅对一个相同签名的函数不算是实现了Getter接口
		除非你额外定义一个结构体

		type myGetter struct {}
		func (g myGetter) Get(key string) ([]byte, error) {
			... // 这里面填充 mygetter的内容
		}
		var g Getter = myGetter{}
		g.Get(...)

		但是这样就很蠢，如果我有mygetter1, mygetter2...岂不是要定义好多个没用的结构体？
		难道不可以直接让一个func对象满足接口吗？
		...
		...
		于是这里的办法就是，定义一个GetFunc类型，表示这样的func
		然后再在这个GetFunc类型上定义一个接口 Get，Get接口内部直接用函数对象来根据参数来调用
		我们只需要把 mygetter转换为 GetFunc类型的对象（当然是可以的！），就可以通过它来赋值给
		Getter接口的对象了

		即
		var g1 Getter = GetFunc(mygetter1)
		var g2 Getter = GetFunc(mygetter2)
		g1.Get(...)
		g2.Get(...)

		Done!

		PS: 这种写法和http包中的HandleFunc类似
		例如:

		helloHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Ni Hao!"))
		}
		http.HandleFunc("/hello", helloHandler)

		在函数内部，进行了
		mux.Handle(pattern, HandlerFunc(handler))
		而这个HandlerFunc，即
		type HandlerFunc func(ResponseWriter, *Request)

		并且再次暴露出ServeHTTP接口
		func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
			f(w, r)
		}
		所以HandlerFunc是对一般http请求处理函数的包装
*/

// Group represents a namespace of cached data
// For example, you can create 3 different groups for student's name, info, and scores
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup creates a new instance of Group. Safe for concurrent calls
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}

	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get fetch a key from local cache. If not stored here, try to load bue.
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[GoCache] cache hit for key = %s", key)
		return v, nil
	}
	log.Printf("[GoCache] cache miss for key = %s, try load", key)
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key) // or remotely
}

func (g *Group) getLocally(key string) (ByteView, error) {
	b, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{
		b: cloneBytes(b),
	}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
