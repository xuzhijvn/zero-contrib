package zookeeper

import (
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/mapping"
)

type target struct {
	Hosts       []string      `key:",optional"`
	BasePath    string        `key:",optional"`
	ServiceName string        `key:",optional"`
	Timeout     time.Duration `key:"timeout,optional"`
}

// parseURL with parameters
func parseURL(u url.URL) (*target, error) {
	tgt := &target{
		BasePath:    "/services",
		ServiceName: u.Path,
		Timeout:     time.Second * 15,
	}

	// 解析hosts
	if u.Host != "" {
		tgt.Hosts = strings.Split(u.Host, ",")
	}

	// 处理路径
	if tgt.ServiceName != "" {
		// 移除开头的斜杠
		tgt.ServiceName = strings.TrimPrefix(tgt.ServiceName, "/")
	}

	// 解析查询参数
	if len(u.RawQuery) > 0 {
		params := make(map[string]interface{}, len(u.Query()))
		for name, values := range u.Query() {
			if len(values) > 0 {
				params[name] = values[0]
			}
		}

		if err := mapping.UnmarshalKey(params, tgt); err != nil {
			return nil, errors.Wrap(err, "unmarshal target error")
		}
	}

	// 验证必要参数
	if len(tgt.Hosts) == 0 {
		return nil, errors.New("hosts cannot be empty")
	}
	if tgt.ServiceName == "" {
		return nil, errors.New("service name cannot be empty")
	}

	return tgt, nil
}

// 解析主机列表字符串，格式为 "host1:port1,host2:port2"
func parseHosts(hosts string) []string {
	hostList := strings.Split(hosts, ",")
	result := make([]string, 0, len(hostList))

	for _, host := range hostList {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}

		// 如果没有指定端口，则添加默认端口2181
		if !strings.Contains(host, ":") {
			host = host + ":2181"
		}

		result = append(result, host)
	}

	return result
}
