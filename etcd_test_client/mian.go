package main

import (
	"fmt"
	"mccache"
)

func main() {
	client := mccache.NewClient("")
	v, err := client.Get("scores", "lisi")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Get successfully v is ", string(v))
}
