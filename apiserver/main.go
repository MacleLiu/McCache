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
	flag.StringVar(&pattern, "pattern", defaultPattern, fmt.Sprintf("The pattern section of the api server URL, default: %s", defaultPattern))
	flag.StringVar(&addr, "addr", defaultApiAddr, fmt.Sprintf("api server's IP address port, default: %s", defaultApiAddr))
	flag.Parse()

	client.StartClient(pattern, addr)
}
