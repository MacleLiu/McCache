package main

import (
	"flag"
	"fmt"
	"log"
	mccache "mccache"
	"os"
	"os/signal"
	"syscall"
)

var db = map[string]string{
	"zhangshan": "104",
	"lisi":      "178",
	"Rex":       "156",
	"Rola":      "139",
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

func main() {
	//创建监听退出chan
	c := make(chan os.Signal, 1)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var addr, etcdAddr string
	flag.StringVar(&addr, "addr", "", "mccache server address")
	flag.StringVar(&etcdAddr, "etcdAddr", "", "etcd server address")
	flag.Parse()
	if addr == "" || etcdAddr == "" {
		fmt.Println("Parameter error")
		return
	}

	mc := createGroup()

	server, _ := mccache.NewServer(addr, etcdAddr)

	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				fmt.Println("Program Exit...", s)
				server.Stop()
				return
			}
		}
	}()

	mc.RegisterServer(server)
	log.Println("mccache is running at", addr)

	server.Start()

}
