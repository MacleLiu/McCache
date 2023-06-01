package mccache

import (
	"context"
	pb "mccache/mccachepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	//连接到远程节点，禁用安全传输，没有加密认证
	conn, err := grpc.Dial(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
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
