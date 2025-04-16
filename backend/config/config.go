package config

import (
	"time"

	"github.com/google/uuid"
)

var Proxy string = ""

// JwtSecretKey 用于签名JWT的密钥
var JwtSecretKey []byte

// 本地镜像存储路径
var LocalImagePath string = "./storage"

// 场景默认过期时间
var DefaultExpirationTime time.Duration = 30 * time.Minute

func init() {
	// 初始化密钥
	refreshJwtKey()

	// 启动定时任务，每天凌晨2点刷新密钥
	go func() {
		for {
			now := time.Now()
			next := now.AddDate(0, 0, 1)
			next = time.Date(next.Year(), next.Month(), next.Day(), 2, 0, 0, 0, next.Location())
			t := time.NewTimer(next.Sub(now))
			<-t.C
			refreshJwtKey()
		}
	}()
}

// refreshJwtKey 刷新JWT密钥
func refreshJwtKey() {
	JwtSecretKey = []byte(uuid.New().String())
}
