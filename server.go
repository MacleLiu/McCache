package mccache

import (
	"context"
	"errors"
	"fmt"
	"log"
	pb "mccache/mccachepb"
	"mccache/register"
	"net"
	"sync"

	"google.golang.org/grpc"
)

const (
	defaultAddr     = "127.0.0.1:9122"
	defaultReplicas = 50
	serviceName     = "mccache"
)

type server struct {
	pb.UnimplementedMcCacheServer

	addr       string //服务器地址 format: ip:port
	mu         sync.Mutex
	grpcServer *grpc.Server
	//etcdRegister *register.EtcdRegister
	// peers *consistenthash.Map //一致性哈希
	// clients map[string]*client
	status bool //服务是否启动

	etcdAddr string //etcd注册中心地址
	//stop chan struct{} //服务停止信号，通知注册器当前服务停止，立即从注册中心移除该节点
}

var _ PeerPicker = (*server)(nil)

// 服务节点日志
func (s *server) Log(format string, v ...interface{}) {
	log.Printf("[McCache-Server %s] %s", s.addr, fmt.Sprintf(format, v...))
}

// NewServer 创建cache的server实例；若addr为空，则使用defaultAddr
func NewServer(addr, etcdAddr string) (*server, error) {
	//创建grpc服务
	grpcServer := grpc.NewServer()
	if addr == "" {
		addr = defaultAddr
	}
	if etcdAddr == "" {
		return nil, fmt.Errorf("etcdAddr is empty")
	}
	server := &server{
		addr:       addr,
		grpcServer: grpcServer,
		etcdAddr:   etcdAddr,
	}
	return server, nil
}

func (s *server) Get(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	//获取客户端发送的参数
	group, key := req.GetGroup(), req.GetKey()
	resp := &pb.Response{}

	s.Log("Receive gRPC Request {%s: %s}", group, key)
	if key == "" {
		return resp, errors.New("key is empty")
	}

	//根据group，获取相应的缓存组实例
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("not funt group[%s]", group)
	}

	//根据key，获取相应的缓存值
	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}

	//将缓存值写入响应体
	resp.Value = view.ByteSlice()
	return resp, nil
}

// 由基于一致性哈希的自定义gRPC负载均衡实现,所以服务端不再进行节点选择
// 实现PeerPicker接口，PickPeer 通过键选择一个远程节点
func (s *server) PickPeer(key string) (PeerGetter, bool) {
	/* s.mu.Lock()
	defer s.mu.Unlock()
	if peer := s.peers.Get(key); peer != "" && peer != s.addr {
		s.Log("Pick remote peer '%s' for key<%s>", peer, key)
		return s.clients[peer], true
	} */
	return nil, false
}

// Start启动cache服务
func (s *server) Start() error {
	s.mu.Lock()
	if s.status {
		s.mu.Unlock()
		return errors.New("server already started")
	}
	s.status = true
	//监听端口
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	//创建grpc服务
	//grpcServer := grpc.NewServer()

	//在grpc服务端注册需要提供的服务
	pb.RegisterMcCacheServer(s.grpcServer, s)
	s.mu.Unlock()

	//创建一个etcd注册器
	etcdRegister, err := register.NewEtcdRegister(s.etcdAddr)
	if err != nil {
		log.Println(err)
		return err
	}
	defer etcdRegister.Close()

	//注册服务到etcd
	err = etcdRegister.Register(serviceName, s.addr, 5)
	if err != nil {
		log.Printf("server[%s] register service to etcd failed, error: %v", s.addr, err)
		return err
	}

	//启动服务
	if err := s.grpcServer.Serve(lis); err != nil {
		s.Log("failed to serve")
		return err
	}
	return nil
}

func (s *server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = false
	s.grpcServer.GracefulStop()
}
