package middleware

import (
	"fmt"
	"net"
)

// 测试网络连通
// TCP, UDP
func IsActive(ip string, port int) bool {

	conn, err := net.Dial("TCP", fmt.Sprintf("%s:%v", ip, port))
	defer conn.Close()
	if err != nil {
		return false
	}
	return true
}
