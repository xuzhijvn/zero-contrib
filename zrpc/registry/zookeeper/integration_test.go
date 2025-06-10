//go:build integration
// +build integration

package zookeeper

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

// 此测试需要一个运行中的Zookeeper服务器
// 运行测试: go test -tags=integration

// 示例gRPC服务实现
type greeterServer struct {
	helloworld.UnimplementedGreeterServer
}

func (s *greeterServer) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func TestIntegration(t *testing.T) {
	// 跳过自动测试
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 启动gRPC服务器
	port := 50051
	addr := fmt.Sprintf("localhost:%d", port)
	lis, err := net.Listen("tcp", addr)
	assert.NoError(t, err)

	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &greeterServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("服务器启动失败: %v", err)
		}
	}()
	defer s.Stop()

	// 注册服务到Zookeeper
	opts := NewZookeeperConfig(
		addr,
		WithHosts([]string{"localhost:2181"}),
		WithServiceName("greeter"),
		WithBasePath("/services"),
	)

	err = RegisterService(opts)
	if err != nil {
		t.Logf("注册服务失败: %v", err)
		t.Skip("Zookeeper服务器可能未运行")
	}

	// 等待服务注册完成
	time.Sleep(1 * time.Second)

	// 创建客户端连接
	conn, err := grpc.Dial(
		"zookeeper://localhost:2181/greeter",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		t.Fatalf("无法连接服务: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := helloworld.NewGreeterClient(conn)

	// 调用服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &helloworld.HelloRequest{Name: "测试"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello 测试", resp.GetMessage())
}

// 示例：如何在真实应用中使用Zookeeper注册中心
func ExampleRegisterService() {
	// 注册服务到Zookeeper
	opts := NewZookeeperConfig(
		"localhost:50051",
		WithHosts([]string{"localhost:2181"}),
		WithServiceName("example-service"),
		WithBasePath("/services"),
		WithSessionTimeout(60),
	)

	err := RegisterService(opts)
	if err != nil {
		fmt.Printf("注册服务失败: %v\n", err)
		return
	}

	fmt.Println("服务已注册到Zookeeper")
}

func ExampleBuilder_Build() {
	// 使用Zookeeper解析器连接服务
	conn, err := grpc.Dial(
		"zookeeper://localhost:2181/example-service",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("连接服务失败: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("成功连接到服务")
}
