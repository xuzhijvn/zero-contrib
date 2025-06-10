package zookeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/resolver"
)

// zkResolver 实现了resolver.Resolver接口
type zkResolver struct {
	conn     *zk.Conn
	cc       resolver.ClientConn
	ctx      context.Context
	cancel   context.CancelFunc
	tgt      *target
	addrChan chan []string
	wg       sync.WaitGroup
}

// ResolveNow 触发立即解析
func (r *zkResolver) ResolveNow(o resolver.ResolveNowOptions) {
	// 不需要实现，因为我们有持续的监听
}

// Close 关闭resolver
func (r *zkResolver) Close() {
	r.cancel()
	r.wg.Wait()
	r.conn.Close()
}

// watchService 监听服务节点的变化
func (r *zkResolver) watchService() {
	r.wg.Add(1)
	defer r.wg.Done()

	for {
		servicePath := path.Join(r.tgt.BasePath, r.tgt.ServiceName)

		// 检查服务路径是否存在
		exists, _, ch, err := r.conn.ExistsW(servicePath)
		if err != nil {
			logError("check service path error", err)
			time.Sleep(time.Second)
			continue
		}

		if !exists {
			// 服务路径不存在，等待创建
			select {
			case <-ch:
				continue
			case <-r.ctx.Done():
				return
			}
		}

		// 获取服务节点列表
		children, _, childCh, err := r.conn.ChildrenW(servicePath)
		if err != nil {
			logError("get service nodes error", err)
			time.Sleep(time.Second)
			continue
		}

		// 处理服务节点
		var endpoints []string
		for _, child := range children {
			nodePath := path.Join(servicePath, child)
			data, _, err := r.conn.Get(nodePath)
			if err != nil {
				logError(fmt.Sprintf("get node data error: %s", nodePath), err)
				continue
			}

			var serviceData map[string]interface{}
			if err := json.Unmarshal(data, &serviceData); err != nil {
				logError(fmt.Sprintf("unmarshal node data error: %s", nodePath), err)
				continue
			}

			address, ok := serviceData["address"].(string)
			if !ok {
				continue
			}

			port, ok := serviceData["port"].(float64)
			if !ok {
				continue
			}

			endpoint := fmt.Sprintf("%s:%d", address, int(port))
			endpoints = append(endpoints, endpoint)
		}

		// 发送新的地址列表
		select {
		case r.addrChan <- endpoints:
		case <-r.ctx.Done():
			return
		}

		// 等待子节点变化或上下文取消
		select {
		case <-childCh:
		case <-r.ctx.Done():
			return
		}
	}
}

// populateEndpoints 将从zookeeper获取的地址更新到grpc客户端连接
func populateEndpoints(ctx context.Context, cc resolver.ClientConn, addrChan <-chan []string) {
	for {
		select {
		case addrs := <-addrChan:
			if len(addrs) == 0 {
				continue
			}

			var newAddrs []resolver.Address
			for _, addr := range addrs {
				newAddrs = append(newAddrs, resolver.Address{Addr: addr})
			}

			if err := cc.UpdateState(resolver.State{Addresses: newAddrs}); err != nil {
				logx.Errorf("failed to update state: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func logError(msg string, err error) {
	fmt.Printf("zookeeper resolver error: %s, %v\n", msg, err)
}
