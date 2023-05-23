package mccache

// ByteView是缓存值的抽象抽象数据类型和只读封装。
type ByteView struct {
	b []byte
}

// 返回缓存数据的长度。
// 在 lru.Cache 的实现中，要求被缓存对象必须实现 Value 接口
func (v ByteView) Len() int {
	return len(v.b)
}

// 返回一个字节切片类型的拷贝
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 返回字符串类型的值
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
