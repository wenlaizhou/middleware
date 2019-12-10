package middleware

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
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

func GetHostname() string {
	host, err := ioutil.ReadFile("/etc/hostname")
	if err != nil {
		mLogger.Error(err.Error())
	}
	return strings.TrimSpace(string(host))
}

// 获取本机ip地址
func GetIpByInterface(interfaceNames ...string) (NetDevice, error) {
	res := NetDevice{}
	ins, err := net.Interfaces()
	if err != nil {
		return res, err
	}
	for _, iInterface := range ins {
		inInterfaces := false
		for _, name := range interfaceNames {
			if iInterface.Name == name {
				inInterfaces = true
				break
			}
		}
		if !inInterfaces {
			continue
		}
		res.Name = iInterface.Name
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
