package config

type System struct {
	Port     int    `yaml:"port"`
	UseRedis bool   `yaml:"useRedis"`
	UseKafka bool   `yaml:"useKafka"`
	ServerID string `yaml:"serverID"`
}
