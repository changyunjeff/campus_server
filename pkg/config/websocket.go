package config

import (
	"time"
)

type WebSocket struct {
	HeartbeatTime   int    `yaml:"heartbeatTime"`   // 心跳检测时间(秒)
	ReadBufferSize  int    `yaml:"readBufferSize"`  // 读取缓冲大小
	WriteBufferSize int    `yaml:"writeBufferSize"` // 写入缓冲大小
	Expire          string `yaml:"expire"`          // Redis存储过期时间
}

// GetExpiration 获取过期时间
func (w *WebSocket) GetExpiration() time.Duration {
	duration, err := time.ParseDuration(w.Expire)
	if err != nil {
		return time.Hour * 12 // 默认12小时
	}
	return duration
}
