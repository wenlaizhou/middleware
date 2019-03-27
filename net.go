package middleware

import (
	"fmt"
	"net"
)

const (
	PROTOCOL_TCP = "tcp"
	PROTOCOL_UDP = "udp"
)

// 测试网络连通
// tcp, udp
func IsActive(protocol string, ip string, port int) bool {
	switch protocol {
	case PROTOCOL_TCP:
	case PROTOCOL_UDP:
		conn, err := net.Dial(protocol, fmt.Sprintf("%s:%v", ip, port))
		if err != nil {
			return false
		}
		defer conn.Close()
		return true
	default:
		return false
	}
	return false
}
