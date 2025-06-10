package zookeeper

// ParseHosts 解析主机列表字符串，格式为 "host1:port1,host2:port2"
// 这是一个导出函数，可以被外部包使用
func ParseHosts(hosts string) []string {
	return parseHosts(hosts)
}
