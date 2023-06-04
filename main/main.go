package main

import (
	"flag"
	"fmt"
	"log"
	mccache "mccache"
)

var db = map[string]string{
	"zhangshan": "001",
	"lisi":      "002",
	"Rex":       "003",
	"Rola":      "123",
}

func createGroup() *mccache.Group {
	return mccache.NewGroup("scores", 2<<10, mccache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, mc *mccache.Group) {
	server, _ := mccache.NewServer(addr)
	//server.SetPeers(addrs...)
	mc.RegisterServer(server)
	log.Println("mccache is running at", addr)
	server.Start()
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "", "mccache server address")
	flag.Parse()
	if addr == "" {
		fmt.Println("addr is empty")
		return
	}
	/* addrMap := map[int]string{
		8001: "localhost:8001",
		8002: "localhost:8002",
		8003: "localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	} */

	mc := createGroup()

	startCacheServer(addr, mc)

}
