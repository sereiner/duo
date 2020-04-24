package conf

import (
	"fmt"
	"strings"
)

// 获取mq名称
func GetMqName(address string) (proto string, raddr []string, err error) {
	addrs := strings.Split(address, "://")
	if len(addrs) == 0 || len(addrs) > 2 {
		err = fmt.Errorf("MQ地址配置错误%s，格式:stomp://192.168.0.1:61613", addrs)
	}
	if len(addrs[0]) == 0 {
		err = fmt.Errorf("MQ地址配置错误%s，格式:stomp://192.168.0.1:61613", addrs)
	}
	proto = addrs[0]
	if len(raddr) > 1 {
		raddr = strings.Split(addrs[1], ",")
	}
	return
}
