# mccache
一个简单的分布式缓存实现，主要是学习模仿GroupCache实现。
并在此基础上实现了基于etcd的服务注册与发现，通过自定义grpc负载均衡实现了基于一致性哈希的缓存节点选择。
# 运行一个示例
## 1. 首先需要运行一个etcd服务，作为mccache的注册中心
## 2. 在/mccache/main目录下提供了一个示例main.go

```go
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
```

编译main.go文件\
`go build main.go`\
运行服务\
`./main -addr="ip:port" -etcdAddr="ip:port"`\
eg:![image](https://github.com/MacleLiu/mccache/assets/61642169/c060ea0c-43c3-42d9-8913-9fc85caab0c7)
## 3. 在/mccache/apiserver目录下提供了api服务节点
编译main.go文件\
`go build main.go`\
运行api服务\
`./main -pattern="/mccache" -addr="localhost:1234" -etcdAddr="127.0.0.1:2379"`\
eg:启动完成，成功发现已注册的服务![image](https://github.com/MacleLiu/mccache/assets/61642169/23f271d3-095d-4ddc-b37f-6b948b971037)
## 4. 请求数据
成功获取缓存数据
![image](https://github.com/MacleLiu/mccache/assets/61642169/4b4d102e-734c-4923-bb82-a94ee329e436)

