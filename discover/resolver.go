package discover

import (
	"encoding/json"
	"log"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/resolver"
)

type etcdResolver struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cc         resolver.ClientConn
	etcdClient *clientv3.Client
	scheme     string
	ipPool     sync.Map
}

func (e *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	log.Println("etcd resolver resolve now")
}

func (e *etcdResolver) Close() {
	log.Println("etcd resolver close")
	e.cancel()
}

func (e *etcdResolver) watcher() {
	watchChan := e.etcdClient.Watch(context.Background(), "/"+e.scheme, clientv3.WithPrefix())

	for {
		select {
		case val := <-watchChan:
			for _, event := range val.Events {
				switch event.Type {
				case 0: //0是有数据增加
					var valueMap = make(map[string]any)
					err := json.Unmarshal(event.Kv.Value, &valueMap)
					if err != nil {
						log.Println("json unmarshal failed")
					}
					e.store(event.Kv.Key, valueMap["Addr"].(string))
					log.Println("put: ", string(event.Kv.Key))
					e.updateState()
				case 1: //1是有数据减少
					log.Println("del: ", string(event.Kv.Value))
					e.del(event.Kv.Key)
					e.updateState()
				}
			}
		case <-e.ctx.Done():
			return
		}
	}
}

func (e *etcdResolver) store(k []byte, v string) {
	e.ipPool.Store(string(k), v)
}

func (s *etcdResolver) del(key []byte) {
	s.ipPool.Delete(string(key))
}

func (e *etcdResolver) updateState() {
	var addrList resolver.State
	e.ipPool.Range(func(k, v any) bool {
		tA, ok := v.(string)
		if !ok {
			return false
		}
		log.Printf("conn.UpdateState key[%v]; val[%v]\n", k, v)
		addrList.Addresses = append(addrList.Addresses, resolver.Address{Addr: tA})
		return true
	})
	e.cc.UpdateState(addrList)
}
