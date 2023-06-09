// register包实现基于etcd的服务注册
package register

import (
	"context"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

// Etcd服务注册器
type EtcdRegister struct {
	etcdCli *clientv3.Client //etcd客户端
	leaseId clientv3.LeaseID //租约ID
	ctx     context.Context
	cancel  context.CancelFunc
}

// 创建租约
func (s *EtcdRegister) CreateLease(expire int64) error {
	res, err := s.etcdCli.Grant(s.ctx, expire)
	if err != nil {
		log.Printf("createLease fail, error: %v\n", err)
		return err
	}

	s.leaseId = res.ID
	return nil
}

// 续租，发送心跳确定服务正常
func (s *EtcdRegister) KeepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	resChan, err := s.etcdCli.KeepAlive(s.ctx, s.leaseId)
	if err != nil {
		log.Printf("keepAlive failed, error: %v\n", err)
		return resChan, err
	}
	return resChan, nil
}

func (s *EtcdRegister) Watcher(serviceName, addr string, resChan <-chan *clientv3.LeaseKeepAliveResponse) {
	for {
		select {
		case <-resChan:
			//log.Printf("续约成功，%v\n", l)
		case <-s.ctx.Done():
			log.Printf("租约关闭")
			return
		}
	}
}

func (s *EtcdRegister) Close() error {
	log.Printf("Close...\n")
	//撤销租约
	s.etcdCli.Revoke(s.ctx, s.leaseId)

	s.cancel()

	return s.etcdCli.Close()
}

/* func (s *EtcdRegister) close() error {
	log.Printf("close...\n")

	//撤销租约
	s.etcdCli.Revoke(s.ctx, s.leaseId)
	log.Printf("租约关闭")
	s.etcdDelete(serviceName, addr)
	log.Printf("节点已删除")
	return s.etcdCli.Close()
} */

// 实例化一个etcd服务注册器
func NewEtcdRegister(etcdAddr string) (*EtcdRegister, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("new etcd client failed, error: %v\n", err)
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())

	svr := &EtcdRegister{
		etcdCli: client,
		ctx:     ctx,
		cancel:  cancelFunc,
	}

	return svr, nil
}

// 注册服务
func (s *EtcdRegister) Register(serviceName, addr string, expire int64) (err error) {
	//创建租约
	err = s.CreateLease(expire)
	if err != nil {
		return
	}

	//在租赁模式注册一个节点
	err = s.etcdAdd(serviceName, addr)
	if err != nil {
		return
	}

	//续租，服务保活
	keepAliveChan, err := s.KeepAlive()
	if err != nil {
		return err
	}

	//监视租约
	go s.Watcher(serviceName, addr, keepAliveChan)

	return nil
}

// etcdAdd 向租约注册一个端点
func (s *EtcdRegister) etcdAdd(serviceName, addr string) error {
	em, err := endpoints.NewManager(s.etcdCli, serviceName)
	if err != nil {
		return nil
	}
	return em.AddEndpoint(context.TODO(), serviceName+"/"+addr, endpoints.Endpoint{Addr: addr}, clientv3.WithLease(s.leaseId))
}

/*
// etcdDelete从服务中删除一个端点
func (s *EtcdRegister) etcdDel(serviceName, addr string) error {
	em, err := endpoints.NewManager(s.etcdCli, serviceName)
	if err != nil {
		return nil
	}
	return em.DeleteEndpoint(context.TODO(), serviceName+"/"+addr)
}
*/
