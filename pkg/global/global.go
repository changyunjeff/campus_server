package global

import (
	"campus2/pkg/config"
	"github.com/IBM/sarama"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	GVA_VIPER  *viper.Viper
	GVA_CONFIG config.Config
	GVA_LOG    *logrus.Logger
	GVA_DB     *gorm.DB
	GVA_REDIS  redis.UniversalClient
	GVA_PRDER  sarama.SyncProducer
	GVA_CSMER  sarama.ConsumerGroup
)
