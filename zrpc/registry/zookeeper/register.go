package zookeeper

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"
	"github.com/zeromicro/go-zero/core/proc"
)

// zkConn 定义了Zookeeper连接所需的接口
type zkConn interface {
	Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error)
	Exists(path string) (bool, *zk.Stat, error)
	Delete(path string, version int32) error
	Close()
}

// RegisterService register service to zookeeper
func RegisterService(opts *Options) error {
	if len(opts.Hosts) == 0 {
		return fmt.Errorf("hosts cannot be empty")
	}
	if len(opts.ServiceName) == 0 {
		return fmt.Errorf("service name cannot be empty")
	}

	pubListenOn := figureOutListenOn(opts.ListenOn)

	host, ports, err := net.SplitHostPort(pubListenOn)
	if err != nil {
		return fmt.Errorf("failed parsing address error: %v", err)
	}
	port, _ := strconv.ParseUint(ports, 10, 16)

	// 连接到zookeeper
	conn, _, err := zk.Connect(opts.Hosts, time.Duration(opts.SessionTimeout)*time.Second)
	if err != nil {
		return fmt.Errorf("connect to zookeeper error: %v", err)
	}

	// 创建基础路径
	servicePath := path.Join(opts.BasePath, opts.ServiceName)
	err = createPath(conn, opts.BasePath)
	if err != nil {
		return fmt.Errorf("create base path error: %v", err)
	}

	err = createPath(conn, servicePath)
	if err != nil {
		return fmt.Errorf("create service path error: %v", err)
	}

	// 服务节点的名称
	serviceID := fmt.Sprintf("%s-%s-%d", opts.ServiceName, host, port)
	serviceNode := path.Join(servicePath, serviceID)

	// 准备服务数据
	serviceData := map[string]interface{}{
		"id":       serviceID,
		"name":     opts.ServiceName,
		"address":  host,
		"port":     port,
		"weight":   opts.Weight,
		"metadata": opts.Metadata,
		"time":     time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(serviceData)
	if err != nil {
		return fmt.Errorf("marshal service data error: %v", err)
	}

	// 创建临时节点
	_, err = conn.Create(serviceNode, data, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("create service node error: %v", err)
	}

	logx.Infof("register service to zookeeper, service: %s, node: %s", opts.ServiceName, serviceNode)

	// 服务注销
	proc.AddShutdownListener(func() {
		if conn != nil {
			err := conn.Delete(serviceNode, -1)
			if err != nil && err != zk.ErrNoNode {
				logx.Info("deregister service error: ", err.Error())
			}
			conn.Close()
			logx.Info("deregistered service from zookeeper server.")
		}
	})

	return nil
}

// 创建路径（如果不存在）
func createPath(conn zkConn, path string) error {
	exists, _, err := conn.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		_, err = conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != allEths {
		return listenOn
	}

	ip := os.Getenv(envPodIP)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
