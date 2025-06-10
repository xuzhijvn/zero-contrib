# Zookeeper注册中心示例

这个示例演示了如何使用Zookeeper作为go-zero的服务注册与发现中心。

## 前提条件

- 安装并运行Zookeeper服务器
- 安装Go 1.16+

## 构建示例

```bash
go build -o zk-example
```

## 运行服务器

```bash
./zk-example --port 50051 --zk localhost:2181 --service greeter
```

这将启动一个gRPC服务器，并将其注册到Zookeeper。

## 运行客户端

```bash
./zk-example --client --zk localhost:2181 --service greeter World
```

这将连接到Zookeeper，发现服务，并发送一个gRPC请求。

## 参数说明

- `--port`: 服务器监听端口（默认：50051）
- `--client`: 运行客户端模式
- `--zk`: Zookeeper服务器地址，多个地址用逗号分隔（默认：localhost:2181）
- `--service`: 服务名称（默认：greeter）

## 注意事项

- 确保Zookeeper服务器正在运行
- 如果在Docker中运行，请确保正确设置网络配置 