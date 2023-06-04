package loadbalancer

import (
	"context"
	"fmt"
	"log"
	"mccache/consistenthash"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
)

type consistentHashPickerBuilder struct{}

func InitConsistentHashBuilder() {
	balancer.Register(
		base.NewBalancerBuilder(
			"consistentHash",
			&consistentHashPickerBuilder{},
			base.Config{HealthCheck: true},
		),
	)
}

func (b *consistentHashPickerBuilder) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	grpclog.Infof("consistentHashPicker: newPicker called with buildInfo: %v", buildInfo)
	if len(buildInfo.ReadySCs) == 0 {
		fmt.Println("===================================")
		fmt.Println("len(buildInfo.ReadySCs) == 0 ")
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	/* // 构造 consistentHashPicker
	picker := &consistentHashPicker{
		subConns:          make(map[string]balancer.SubConn),
		hash:              consistenthash.New(50, nil), // 构造一致性hash
		consistentHashKey: b.consistentHashKey,         // 用于计算hash的key
	}

	for sc, conInfo := range buildInfo.ReadySCs {
		fmt.Println("+++++++++++++++++++++++")
		fmt.Println("vvvvvvvvvvvvvvvvvvvvvvvvvvv")
		fmt.Println(sc)
		fmt.Println("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv")
		fmt.Println(conInfo)

		node := conInfo.Address.Addr
		picker.hash.Add(node)
		picker.subConns[node] = sc
	}
	return picker */
	subConns := make(map[string]balancer.SubConn)
	for sc, conInfo := range buildInfo.ReadySCs {
		subConns[conInfo.Address.Addr] = sc
	}

	return NewConsistentHashPicker(subConns)
}

type consistentHashPicker struct {
	subConns   map[string]balancer.SubConn
	hash       *consistenthash.Map
	needReport bool
	reportChan chan<- PickResult
}

type PickResult struct {
	Ctx context.Context
	SC  balancer.SubConn
}

func NewConsistentHashPicker(subConns map[string]balancer.SubConn) *consistentHashPicker {
	addrs := make([]string, 0)
	for addr := range subConns {
		addrs = append(addrs, addr)
	}
	log.Printf("consistent hash picker built with addresses %v\n", addrs)
	picker := &consistentHashPicker{
		subConns:   subConns,
		hash:       consistenthash.New(50, nil), // 构造一致性hash
		needReport: false,
	}
	picker.hash.Add(addrs...)
	return picker
}

func NewConsistentHashPickerWithReportChan(subConns map[string]balancer.SubConn, reportChan chan<- PickResult) *consistentHashPicker {
	addrs := make([]string, 0)
	for addr := range subConns {
		addrs = append(addrs, addr)
	}
	log.Printf("consistent hash picker built with addresses %v\n", addrs)
	picker := &consistentHashPicker{
		subConns:   subConns,
		hash:       consistenthash.New(50, nil), // 构造一致性hash
		needReport: true,
		reportChan: reportChan,
	}
	picker.hash.Add(addrs...)
	return picker
}

func (p *consistentHashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var ret balancer.PickResult
	key, ok := info.Ctx.Value(Key).(string)
	if !ok {
		key = info.FullMethodName
	}
	log.Printf("pick for %s\n", key)
	if targetAddr, ok := p.hash.Get(key); ok {
		ret.SubConn = p.subConns[targetAddr]
		if p.needReport {
			p.reportChan <- PickResult{Ctx: info.Ctx, SC: ret.SubConn}
		}
	}
	//ret.SubConn = p.subConns["localhost:50000"]
	if ret.SubConn == nil {
		return ret, balancer.ErrNoSubConnAvailable
	}
	return ret, nil
}

/* func (p *consistentHashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var ret balancer.PickResult
	//key, ok := info.Ctx.Value(p.consistentHashKey).(string)
	fmt.Println("我是目标哈希key", p.consistentHashKey)
	//if ok {
	targetAddr := p.hash.Get(p.consistentHashKey) // 根据key的hash值挑选出对应的节点

	ret.SubConn = p.subConns[targetAddr]
	//}
	return ret, nil
} */
