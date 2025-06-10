package zookeeper

import (
	"context"
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
	"google.golang.org/grpc/resolver"
)

const (
	// schemeName for the urls
	// All target URLs like 'zookeeper://.../...' will be resolved by this resolver
	schemeName = "zookeeper"
)

func init() {
	resolver.Register(&builder{})
}

// builder implements resolver.Builder and use for constructing all zookeeper resolvers
type builder struct{}

func (b *builder) Build(url resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	tgt, err := parseURL(url.URL)
	if err != nil {
		return nil, fmt.Errorf("wrong zookeeper URL: %v", err)
	}

	// 连接到zookeeper
	conn, _, err := zk.Connect(tgt.Hosts, time.Duration(tgt.Timeout)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to the zookeeper: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &zkResolver{
		conn:     conn,
		cc:       cc,
		ctx:      ctx,
		cancel:   cancel,
		tgt:      tgt,
		addrChan: make(chan []string, 1),
	}

	// 启动监听服务节点变化的goroutine
	go r.watchService()
	// 启动更新服务地址的goroutine
	go populateEndpoints(ctx, cc, r.addrChan)

	return r, nil
}

// Scheme returns the scheme supported by this resolver.
func (b *builder) Scheme() string {
	return schemeName
}
