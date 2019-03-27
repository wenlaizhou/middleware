package middleware

import (
	"fmt"
	"net"
)

// 测试网络连通
// TCP, UDP
func IsActive(ip string, port int) bool {

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%v", ip, port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
