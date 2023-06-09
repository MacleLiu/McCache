package client

import (
	"context"
	"log"
	"mccache/client/discover"
	"mccache/loadbalancer"
	pb "mccache/mccachepb"
	"mccache/singleflight"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type client struct {
	pattern  string
	addr     string
	etcdAddr string
	loader   *singleflight.Group
}

func NewClient(pattern, addr, etcdAddr string) *client {
	return &client{
		pattern:  pattern,
		addr:     addr,
		etcdAddr: etcdAddr,
		loader:   &singleflight.Group{},
	}
}

func (c *client) Start() {
	//初始化balancer
	loadbalancer.InitConsistentHashBuilder()

	//注册自定义etcd解析器
	etcdResolverBuilder := discover.NewEtcdResolverBuilder(c.etcdAddr)
	resolver.Register(etcdResolverBuilder)

	// 使用自带的DNS解析器和负载均衡实现方式
	conn, err := grpc.Dial(
		"etcd:///mccache",
		grpc.WithDefaultServiceConfig(`{"LoadBalancingPolicy": "consistentHash"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 创建gPRC客户端
	grpcClient := pb.NewMcCacheClient(conn)

	http.Handle(c.pattern, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			group := r.URL.Query().Get("group")

			//通过singleflight执行RPC调用
			resp, err := c.loader.Do(group, key, func() (any, error) {
				if resp, err := grpcClient.Get(context.WithValue(context.Background(), loadbalancer.Key, key), &pb.Request{
					Group: group,
					Key:   key,
				}); err == nil {
					return resp, nil
				} else {
					log.Printf("grpc call for {group: %s, key: %s} failed. Error %v", group, key, err)
					return nil, err
				}
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(resp.(*pb.Response).GetValue())

		}))
	log.Println("fontend server is running at", c.addr)
	log.Fatal(http.ListenAndServe(c.addr, nil))
}
