package main

import (
	"fmt"
	"log"
	"net/http"

	mccache "McCache"
)

var db = map[string]string{
	"zhangshan": "001",
	"lisi":      "002",
	"Rex":       "003",
	"Rola":      "123",
}

func main() {
	mccache.NewGroup("stuNum", 2<<10, mccache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB search key]", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := mccache.NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
