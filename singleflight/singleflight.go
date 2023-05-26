// singleflight包，使多个相同请求等待第一个请求，来防止缓存击穿
package singleflight

import "sync"

// call是正在进行或已结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 是主数据结构，管理不同key的请求
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	//如果已经存在对该key的请求，等待请求结束，返回结果
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	//新的key的请求
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c //添加到map中，表示该key已经存在对应的请求
	g.mu.Unlock()

	c.val, c.err = fn() //调用fn，发起请求
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key) //请求结束，更新g.m
	g.mu.Unlock()

	return c.val, c.err
}
