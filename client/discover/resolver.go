package discover

import (
	"encoding/json"
	"log"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"golang.org/x/net/context"
	"google.golang.org/grpc/resolver"
)

type etcdResolver struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cc         resolver.ClientConn
	etcdClient *clientv3.Client
	scheme     string
	nodePool   sync.Map //客户端从etcd发现的服务节点信息
}

func (e *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	log.Println("etcd resolver resolve now")
}

func (e *etcdResolver) Close() {
	log.Println("etcd resolver close")
	e.cancel()
}

func (e *etcdResolver) watcher() {
	log.Println("watching......")
	defer log.Println("watching end......")
	watchChan := e.etcdClient.Watch(context.Background(), "mccache", clientv3.WithPrefix())

	for {
		select {
		case val := <-watchChan:
			for _, event := range val.Events {
				switch event.Type {
				case 0: //0是有数据增加
					/* var valueMap = make(map[string]any)
					err := json.Unmarshal(event.Kv.Value, &valueMap)
					if err != nil {
						log.Println("json unmarshal failed")
					} */
					e.store(event.Kv.Key, event.Kv.Value)
					log.Println("put: ", string(event.Kv.Key))
					e.updateState()
				case 1: //1是有数据减少
					log.Println("del: ", string(event.Kv.Key))
					e.del(event.Kv.Key)
					e.updateState()
				}
			}
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *etcdResolver) store(k, v []byte) {
	e.nodePool.Store(string(k), v)
}

func (s *etcdResolver) del(key []byte) {
	s.nodePool.Delete(string(key))
}

func (e *etcdResolver) updateState() {
	var addrList resolver.State
	e.nodePool.Range(func(k, v any) bool {
		nodeInfo, ok := v.([]byte)
		if !ok {
			return false
		}

		var endPoint = new(endpoints.Endpoint)
		err := json.Unmarshal(nodeInfo, endPoint)
		if err != nil {
			log.Println("json unmarshal failed")
		}

		log.Printf("conn.UpdateState key[%v]; val[%v]\n", k, endPoint.Addr)
		addrList.Addresses = append(addrList.Addresses, resolver.Address{Addr: endPoint.Addr})
		return true
	})

	e.cc.UpdateState(addrList)
}
