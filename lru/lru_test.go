package lru

import "testing"

type String string

func (d String) Len() int {
	return len(d)
}

var testData = []struct {
	key   string
	value String
}{
	{"key1", String("1234")},
	{"key2", String("1a2b")},
	{"key3", String("abcd")},
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	for _, data := range testData {
		lru.Add(data.key, data.value)
	}
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if v, ok := lru.Get("key2"); !ok || string(v.(String)) != "1a2b" {
		t.Fatalf("cache hit key2=1a2b failed")
	}
	if v, ok := lru.Get("key3"); !ok || string(v.(String)) != "abcd" {
		t.Fatalf("cache hit key3=abcd failed")
	}
	if _, ok := lru.Get("key4"); ok {
		t.Fatalf("cache miss key4 failed")
	}
}

func TestRemove(t *testing.T) {
	lru := New(int64(0), nil)
	for _, data := range testData {
		lru.Add(data.key, data.value)
	}
	lru.Remove("key2")
	if _, ok := lru.Get("key2"); ok || lru.Len() != 2 {
		t.Fatal("Remove key2 failed")
	}
	lru.Remove("key1")
	if _, ok := lru.Get("key1"); ok || lru.Len() != 1 {
		t.Fatal("Remove key1 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	lru := New(int64(20), nil)
	for _, data := range testData {
		lru.Add(data.key, data.value)
	}
	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("RemoveOldest key1 failed")
	}
	if v, ok := lru.Get("key3"); !ok || string(v.(String)) != "abcd" {
		t.Fatalf("cache hit key3=abcd failed")
	}
}
