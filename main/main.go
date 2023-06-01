package main

import (
	"flag"
	"fmt"
	"log"
	mccache "mccache"
	"net/http"
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

func startCacheServer(addr string, addrs []string, mc *mccache.Group) {
	server, _ := mccache.NewServer(addr)
	server.SetPeers(addrs...)
	mc.RegisterServer(server)
	log.Println("mccache is running at", addr)
	server.Start()
}

func startAPIServer(apiAddr string, mc *mccache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := mc.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "mccache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "localhost:8001",
		8002: "localhost:8002",
		8003: "localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	mc := createGroup()
	if api {
		go startAPIServer(apiAddr, mc)
	}
	startCacheServer(addrMap[port], []string(addrs), mc)
}
