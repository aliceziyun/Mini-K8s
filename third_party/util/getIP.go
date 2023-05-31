package util

import (
	"fmt"
	"net"
)

var IpMap = map[string]string{
	"192.168.1.4":  "10.119.10.175",
	"192.168.1.6":  "10.119.11.91",
	"192.168.1.11": "10.119.11.108",
}

// GetIP :获取内网IP
func GetIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

// GetDynamicIP :获取浮动IP
func GetDynamicIP() string {
	ips, _ := GetIP()
	return IpMap[ips]
}
