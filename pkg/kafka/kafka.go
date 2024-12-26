package kafka

import (
	"campus2/pkg/config"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

var (
	producer     sarama.SyncProducer
	consumer     sarama.ConsumerGroup
	producerOnce sync.Once
	consumerOnce sync.Once
)

// NewKafkaProducer 创建生产者
func NewKafkaProducer(cfg config.Kafka) (sarama.SyncProducer, error) {
	var err error
	producerOnce.Do(func() {
		config := sarama.NewConfig()
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Retry.Max = 5
		config.Producer.Return.Successes = true

		producer, err = sarama.NewSyncProducer(cfg.Brokers, config)
	})
	return producer, err
}

// NewKafkaConsumer 创建消费者组
func NewKafkaConsumer(cfg config.Kafka) (sarama.ConsumerGroup, error) {
	var err error
	consumerOnce.Do(func() {
		config := sarama.NewConfig()
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
		config.Consumer.Offsets.Initial = sarama.OffsetNewest

		consumer, err = sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, config)
	})
	return consumer, err
}

// ConsumerHandler 消费者处理器
type ConsumerHandler struct {
	ready chan bool
	// 可以添加其他需要的字段
}

// Setup 在消费者会话开始时调用
func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup 在消费者会话结束时调用
func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消息的主要方法
func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			log.Printf("Message topic:%q partition:%d offset:%d\n",
				message.Topic, message.Partition, message.Offset)

			// 处理消息
			// TODO: 实现具体的消息处理逻辑

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// StartConsumerGroup 启动消费者组
func StartConsumerGroup(ctx context.Context, cfg config.Kafka, topics []string) error {
	handler := &ConsumerHandler{
		ready: make(chan bool),
	}

	for {
		err := consumer.Consume(ctx, topics, handler)
		if err != nil {
			return fmt.Errorf("Error from consumer: %v", err)
		}

		if ctx.Err() != nil {
			return nil
		}
		handler.ready = make(chan bool)
	}
}

// SendMessage 发送消息
func SendMessage(topic string, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Message sent to partition %d at offset %d\n", partition, offset)
	return nil
}

// Close 关闭生产者和消费者
func Close() {
	if producer != nil {
		producer.Close()
	}
	if consumer != nil {
		consumer.Close()
	}
}
