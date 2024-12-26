package config

import "time"

type Redis struct {
	Addr         string   `yaml:"addr"`
	Password     string   `yaml:"password"`
	DB           int      `yaml:"db"`
	PoolSize     int      `yaml:"poolSize"`
	Duration     string   `yaml:"duration"`     // 默认缓存时间
	UseCluster   bool     `yaml:"useCluster"`   // 是否使用集群模式
	ClusterAddrs []string `yaml:"clusterAddrs"` // 集群节点地址
}

// GetDuration 获取缓存时间
func (r *Redis) GetDuration() time.Duration {
	duration, err := time.ParseDuration(r.Duration)
	if err != nil {
		return time.Hour * 72 // 默认3天
	}
	return duration
}
