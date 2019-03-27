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
func IsActive(protocol string, ip string, port int) bool {
	switch protocol {
	case PROTOCOL_TCP, PROTOCOL_UDP:
		conn, err := net.DialTimeout(protocol, fmt.Sprintf("%s:%v", ip, port), 3*time.Second)
		if err != nil {
			return false
		}
		defer conn.Close()
		return true
	default:
		return false
	}
}
