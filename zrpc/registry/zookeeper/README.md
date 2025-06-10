# Zookeeper Registry for go-zero

这个包提供了基于Zookeeper的服务注册与发现功能，用于go-zero框架。

## 安装

```bash
go get -u github.com/zeromicro/zero-contrib/zrpc/registry/zookeeper
```

## 使用方法

### 服务注册

在你的服务中，你可以使用以下代码注册服务到Zookeeper：

```go
import (
    "github.com/zeromicro/go-zero/zrpc"
    "github.com/zeromicro/zero-contrib/zrpc/registry/zookeeper"
)

func main() {
    server := zrpc.MustNewServer(zrpc.RpcServerConf{
        ListenOn: "127.0.0.1:8080",
        // 其他配置...
    }, func(grpcServer *grpc.Server) {
        // 注册你的服务...
    })
    
    // 配置Zookeeper注册
    opts := zookeeper.NewZookeeperConfig(
        server.RpcServer().Address(),
        zookeeper.WithHosts([]string{"127.0.0.1:2181"}),
        zookeeper.WithServiceName("your-service-name"),
        zookeeper.WithBasePath("/services"),
    )
    
    // 注册服务
    if err := zookeeper.RegisterService(opts); err != nil {
        log.Fatal(err)
    }
    
    server.Start()
}
```

### 服务发现

在客户端，你可以使用以下代码连接到Zookeeper注册的服务：

```go
import (
    "github.com/zeromicro/go-zero/zrpc"
)

func main() {
    // 使用zookeeper://host:port/service-name格式的Target
    client, err := zrpc.NewClient(zrpc.RpcClientConf{
        Target: "zookeeper://127.0.0.1:2181/your-service-name",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 现在可以使用client调用远程服务了
}
```

## 配置选项

### 服务注册选项

- `WithHosts(hosts []string)`: 设置Zookeeper服务器地址列表
- `WithBasePath(basePath string)`: 设置服务注册的基础路径，默认为"/services"
- `WithServiceName(serviceName string)`: 设置服务名称
- `WithWeight(weight float64)`: 设置服务权重
- `WithMetadata(metadata map[string]string)`: 设置服务元数据
- `WithSessionTimeout(timeout int)`: 设置会话超时时间(秒)，默认为60秒
- `WithConnectTimeout(timeout int)`: 设置连接超时时间(秒)，默认为5秒

### 服务发现URL参数

在Target URL中可以使用以下查询参数：

- `timeout`: 连接超时时间(秒)，默认为15秒
- `basePath`: 服务注册的基础路径，默认为"/services"

例如：
```
zookeeper://127.0.0.1:2181/your-service-name?timeout=30&basePath=/my-services
``` 