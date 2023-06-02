// discover包实现基于etcd的服务发现
package discover

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type etcdResolverBuilder struct {
	etcdClient *clientv3.Client
}

func NewEtcdResolverBuilder() *etcdResolverBuilder {
	//创建etcd客户端连接
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("create etcd client failed, error: %v ", err)
		panic(err)
	}

	return &etcdResolverBuilder{
		etcdClient: etcdClient,
	}
}

func (b *etcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// 获取指定前缀的etcd节点值
	//prefix := "/" + target.URL.Scheme
	prefix := "mccache"
	log.Println(prefix)

	// 获取指定前缀的etcd节点值
	res, err := b.etcdClient.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Printf("Build etcd get addr failed, error: %v\n", err)
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())

	es := &etcdResolver{
		cc:         cc,
		etcdClient: b.etcdClient,
		ctx:        ctx,
		cancel:     cancelFunc,
		scheme:     target.URL.Scheme,
	}

	//将获取到的服务地址保存到本地map中
	log.Printf("etcd res:%+v\n", res)
	for _, kv := range res.Kvs {
		var valueMap = make(map[string]any)
		err := json.Unmarshal(kv.Value, &valueMap)
		if err != nil {
			log.Println("json unmarshal failed, error: ", err)
		}
		es.store(kv.Key, valueMap["Addr"].(string))
		fmt.Println("这是保存的值：", string(kv.Key), "--", valueMap["Addr"])
	}

	//更新地址列表
	es.updateState()

	//监视etcd中的服务是否发生变化
	go es.watcher()

	return es, nil

}

func (b *etcdResolverBuilder) Scheme() string {
	return "etcd"
}
