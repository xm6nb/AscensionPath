package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAvailablePort 获取当前系统可用的端口
func GetAvailablePort() (int, error) {
	// 监听一个随机端口
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	// 获取实际分配的端口
	return l.Addr().(*net.TCPAddr).Port, nil
}

// GetLocalIP 获取当前主机的IP地址
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", nil
}

// StringToInt 字符串转整型
func StringToInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func GetDataFromContext[T any](c *gin.Context, key string, datatype T) (T, error) {
	dataInterface, ok := c.Get(key)
	if !ok || dataInterface == nil {
		return datatype, errors.New(key + "数据不存在")
	}
	// 强制类型转换为datatype
	data, ok := dataInterface.(T)
	if !ok {
		return datatype, errors.New(key + "数据类型转换失败")
	}
	return data, nil
}

// MD5Encode 对字符串进行MD5加密
func MD5Encode(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
