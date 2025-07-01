package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zeromicro/zero-contrib/zrpc/registry/zookeeper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

var (
	port        = flag.Int("port", 50051, "服务器端口")
	client      = flag.Bool("client", false, "运行客户端模式")
	zkServers   = flag.String("zk", "tony2c4g:2181", "Zookeeper服务器地址，多个地址用逗号分隔")
	serviceName = flag.String("service", "greeter", "服务名称")
)

// 服务器实现
type greeterServer struct {
	helloworld.UnimplementedGreeterServer
}

func (s *greeterServer) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	log.Printf("收到请求: %v", in.GetName())
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()

	if *client {
		runClient()
	} else {
		runServer()
	}
}

// 运行服务器
func runServer() {
	// 创建监听器
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	// 创建gRPC服务器
	s := grpc.NewServer()
	helloworld.RegisterGreeterServer(s, &greeterServer{})

	// 注册服务到Zookeeper
	addr := fmt.Sprintf("localhost:%d", *port)
	opts := zookeeper.NewZookeeperConfig(
		addr,
		zookeeper.WithHosts(parseZkServers(*zkServers)),
		zookeeper.WithServiceName(*serviceName),
	)

	err = zookeeper.RegisterService(opts)
	if err != nil {
		log.Fatalf("注册服务到Zookeeper失败: %v", err)
	}

	log.Printf("服务器启动在 %v，已注册到Zookeeper", addr)

	// 处理优雅退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("正在关闭服务器...")
		s.GracefulStop()
	}()

	// 启动服务器
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// 运行客户端
func runClient() {
	// 使用Zookeeper解析器连接服务
	target := fmt.Sprintf("zookeeper://%s/%s", *zkServers, *serviceName)
	conn, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		log.Fatalf("连接服务失败: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	client := helloworld.NewGreeterClient(conn)

	// 发送请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := "world"
	//if len(flag.Args()) > 0 {
	//	name = flag.Args()[0]
	//}

	resp, err := client.SayHello(ctx, &helloworld.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	log.Printf("收到响应: %s", resp.GetMessage())
}

// 解析Zookeeper服务器地址
func parseZkServers(servers string) []string {
	return zookeeper.ParseHosts(servers)
}
