package lru

import "container/list"

// LRU缓存结构，并发不安全
type Cache struct {
	maxBytes int64 //最大允许使用内存,0表示不设置最大内存限制
	nbytes   int64 //当前已使用内存
	ll       *list.List
	cache    map[string]*list.Element
	//缓存条目删除时可执行的回调函数
	OnEvicted func(key string, value Value)
}

// 双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// 构造函数，新建一个Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add向缓存中新增或修改值
func (c *Cache) Add(key string, value Value) {
	// 如果键已存在，更新节点的值，并移动到表头
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 超过存储空间上限，移除最少访问的节点
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// Get从缓存中查找一个键值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if c.cache == nil {
		return
	}
	// 从字典中找到对应的双向链表的节点
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 将该节点移动到链表头部
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove 从缓存中移除指定的键
func (c *Cache) Remove(key string) {
	if c.cache == nil {
		return
	}
	if ele, ok := c.cache[key]; ok {
		c.removeElement(ele)
	}
}

// 移除最少访问的节点
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

// remove操作的具体实现
func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len 获取当前数据条数
func (c *Cache) Len() int {
	return c.ll.Len()
}

// Clear 清空缓存
