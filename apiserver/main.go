package main

import (
	"flag"
	"fmt"
	"mccache/client"
)

const (
	defaultApiAddr = "localhost:9999"
	defaultPattern = "/api"
)

func main() {
	var pattern string
	var addr string
	var etcdAddr string
	flag.StringVar(&pattern, "pattern", defaultPattern, fmt.Sprintf("The pattern section of the api server URL, default: %s", defaultPattern))
	flag.StringVar(&addr, "addr", defaultApiAddr, fmt.Sprintf("api server's IP address port, default: %s", defaultApiAddr))
	flag.StringVar(&etcdAddr, "etcdAddr", "", "etcd server address")
	flag.Parse()
	if pattern == "" || addr == "" || etcdAddr == "" {
		fmt.Println("Parameter error")
		return
	}

	client := client.NewClient(pattern, addr, etcdAddr)
	client.Start()
}
