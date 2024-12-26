package config

import (
	"time"
)

type Kafka struct {
	Brokers           []string `yaml:"brokers"`           // Kafka代理地址
	ConsumerGroup     string   `yaml:"consumerGroup"`     // 消费者组ID
	Topic             string   `yaml:"topic"`             // 主题
	MessageExpiration string   `yaml:"messageExpiration"` // 消息过期时间
}

// GetMessageExpiration 获取消息过期时间
func (k *Kafka) GetMessageExpiration() time.Duration {
	duration, err := time.ParseDuration(k.MessageExpiration)
	if err != nil {
		return time.Hour * 24 // 默认24小时
	}
	return duration
}
