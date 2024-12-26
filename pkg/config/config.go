package config

type Config struct {
	System    System    `yaml:"system"`
	Mysql     Mysql     `yaml:"mysql"`
	Logrus    Logrus    `yaml:"logrus"`
	Redis     Redis     `yaml:"redis"`
	WebSocket WebSocket `yaml:"websocket"`
	Kafka     Kafka     `yaml:"kafka"`
}
