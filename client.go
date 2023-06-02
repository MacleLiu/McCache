package mccache

import (
	"context"
	"fmt"
	"mccache/discover"
	pb "mccache/mccachepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type client struct {
	addr string // 目标节点地址 ip:addr
}

var _ PeerGetter = (*client)(nil)

func NewClient(addr string) *client {
	return &client{addr: addr}
}

// 实现PeerGetter接口
func (c *client) Get(group string, key string) ([]byte, error) {
	//注册自定义etcd解析器
	etcdResolverBuilder := discover.NewEtcdResolverBuilder()
	resolver.Register(etcdResolverBuilder)

	// 使用自带的DNS解析器和负载均衡实现方式
	conn, err := grpc.Dial("etcd:///mccache", grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	/* //连接到远程节点，禁用安全传输，没有加密认证
	conn, err := grpc.Dial(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	} */
	defer conn.Close()

	//建立连接
	grpcClient := pb.NewMcCacheClient(conn)

	//执行RPC调用
	resp, err := grpcClient.Get(context.Background(), &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetValue(), nil
}
