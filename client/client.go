package client

import (
	"context"
	"log"
	"mccache/client/discover"
	"mccache/loadbalancer"
	pb "mccache/mccachepb"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

/* type client struct {
	addr string // 目标节点地址 ip:addr
}

func NewClient(addr string) *client {
	return &client{addr: addr}
} */

func StartClient(pattern, addr string) {
	//初始化balancer
	loadbalancer.InitConsistentHashBuilder()

	//注册自定义etcd解析器
	etcdResolverBuilder := discover.NewEtcdResolverBuilder()
	resolver.Register(etcdResolverBuilder)

	// 使用自带的DNS解析器和负载均衡实现方式
	conn, err := grpc.Dial(
		"etcd:///mccache",
		grpc.WithDefaultServiceConfig(`{"LoadBalancingPolicy": "consistentHash"}`),
		//grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 创建gPRC客户端
	grpcClient := pb.NewMcCacheClient(conn)

	http.Handle(pattern, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			group := r.URL.Query().Get("group")

			//执行RPC调用
			resp, err := grpcClient.Get(context.WithValue(context.Background(), loadbalancer.Key, key), &pb.Request{
				Group: group,
				Key:   key,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(resp.GetValue())

		}))
	log.Println("fontend server is running at", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
