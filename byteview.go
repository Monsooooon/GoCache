package gocache

// ByteView is an abstraction of a continuous memory space of type byte
// It implements the Value interface, therefore can be used in lru
// content of ByteView should be read-only to outer program
type ByteView struct {
	b []byte
}

/* 由于 ByteView 只包含了一个slice， 而slice本身就是一个很小的结构体
   因此用指针类型还是原类型作为接收参数都是可以的
*/

// ByteSlice returns a copy of byte data
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}

// Len returns the size of used memory space
func (bv ByteView) Len() int64 {
	return int64(len(bv.b))
}

// String prints the content of this byte slice
func (bv ByteView) String() string {
	return string(bv.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
