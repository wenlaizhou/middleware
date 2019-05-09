package middleware

import (
	"fmt"
	"net"
	"time"
)

const (
	PROTOCOL_TCP = "tcp"
	PROTOCOL_UDP = "udp"
)

// 测试网络连通
// tcp, udp
func IsActive(protocol string, ip string, port int, timeoutSecond int) bool {
	switch protocol {
	case PROTOCOL_TCP, PROTOCOL_UDP:
		conn, err := net.DialTimeout(protocol, fmt.Sprintf("%s:%v", ip, port), time.Duration(timeoutSecond)*time.Second)
		if err != nil {
			return false
		}
		defer conn.Close()
		return true
	default:
		return false
	}
}

// 获取本机ip地址
func GetIpAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
