package init

import (
	"campus2/pkg"
	"campus2/pkg/global"
	"campus2/pkg/kafka"
	"campus2/pkg/redis"
	"context"
	"fmt"
)

func init() {
	global.GVA_VIPER = pkg.NewViper()
	fmt.Println("Viper 初始化完成 现在可以使用global.GVA_CONFIG来访问配置文件的内容了")
	global.GVA_LOG = pkg.NewLogrus(context.Background())
	fmt.Println("Logrus 初始化完成")

	// 只有在配置文件中启用Redis时才初始化
	if global.GVA_CONFIG.System.UseRedis {
		global.GVA_REDIS = redis.GetRedis(global.GVA_CONFIG.Redis)
		fmt.Println("Redis 初始化完成")
	}

	if global.GVA_CONFIG.System.UseKafka {
		// 初始化Kafka
		if producer, err := kafka.NewKafkaProducer(global.GVA_CONFIG.Kafka); err != nil {
			global.GVA_LOG.Fatalf("Failed to create Kafka producer: %v", err)
		} else {
			global.GVA_PRDER = producer
		}

		if consumer, err := kafka.NewKafkaConsumer(global.GVA_CONFIG.Kafka); err != nil {
			global.GVA_LOG.Fatalf("Failed to create Kafka consumer: %v", err)
			global.GVA_CSMER = consumer
		}

		// 启动Kafka消费者组
		go func() {
			topics := []string{global.GVA_CONFIG.Kafka.Topic}
			if err := kafka.StartConsumerGroup(context.Background(), global.GVA_CONFIG.Kafka, topics); err != nil {
				global.GVA_LOG.Errorf("Kafka consumer error: %v", err)
			}
		}()
	}

}
