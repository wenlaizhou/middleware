package middleware

import (
	"errors"
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

type NetDevice struct {
	Name string
	Ip   string
	Mac  string
}

// 获取本机ip地址
func GetIpByInterface(name string) (NetDevice, error) {
	res := NetDevice{}
	res.Name = name
	ins, err := net.Interfaces()
	if err != nil {
		return res, err
	}
	for _, iInterface := range ins {
		if iInterface.Name != name {
			continue
		}
		res.Mac = iInterface.HardwareAddr.String()
		addrs, err := iInterface.Addrs()
		if err != nil {
			return res, err
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					res.Ip = ipnet.IP.String()
					return res, nil
				}
			}
		}
	}
	return res, errors.New("no this device")
}
