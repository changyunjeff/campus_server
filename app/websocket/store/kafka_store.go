package store

import (
	"campus2/app/websocket/model"
	"campus2/pkg/global"
	"campus2/pkg/kafka"
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaMessageStore struct {
	topic string
}

func NewKafkaMessageStore(topic string) *KafkaMessageStore {
	return &KafkaMessageStore{
		topic: topic,
	}
}

// StoreMessage 存储离线消息
func (s *KafkaMessageStore) StoreMessage(msg *model.OfflineMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		global.GVA_LOG.Errorf("序列化消息失败: %v", err)
		return err
	}

	// 使用kafka包提供的SendMessage方法
	err = kafka.SendMessage(s.topic, msg.To, data)
	if err != nil {
		global.GVA_LOG.Errorf("存储离线消息到Kafka失败: %v", err)
		return err
	}

	global.GVA_LOG.Infof("离线消息已存储到Kafka: topic=%s, userID=%s", s.topic, msg.To)
	return nil
}

// SendMessage 发送消息到指定topic
func (s *KafkaMessageStore) SendMessage(topic string, key string, message []byte) error {
	global.GVA_LOG.Infof("发送消息到Kafka, topic: %s, key: %s", topic, key)

	err := kafka.SendMessage(topic, key, message)
	if err != nil {
		global.GVA_LOG.Errorf("发送消息到Kafka失败: topic=%s, key=%s, error=%v", topic, key, err)
		return fmt.Errorf("send message failed: %v", err)
	}

	global.GVA_LOG.Infof("消息发送成功: topic=%s, key=%s", topic, key)
	return nil
}

// GetOfflineMessages 获取离线消息
func (s *KafkaMessageStore) GetOfflineMessages(userID string) ([]*model.OfflineMessage, error) {
	global.GVA_LOG.Infof("开始获取用户 %s 的离线消息", userID)

	// TODO: 实现从Kafka消费消息的逻辑
	// 1. 创建一个专门的消费者处理器来处理离线消息
	handler := &OfflineMessageHandler{
		userID:   userID,
		messages: make([]*model.OfflineMessage, 0),
	}

	// 2. 使用消费者组消费消息
	ctx := context.Background()
	err := global.GVA_CSMER.Consume(ctx, []string{s.topic}, handler)
	if err != nil {
		global.GVA_LOG.Errorf("消费离线消息失败: %v", err)
		return nil, err
	}

	return handler.messages, nil
}

// OfflineMessageHandler 离线消息处理器
type OfflineMessageHandler struct {
	userID   string
	messages []*model.OfflineMessage
	ready    chan bool
}

func (h *OfflineMessageHandler) Setup(_ sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *OfflineMessageHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *OfflineMessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// 只处理属于指定用户的消息
		if string(message.Key) == h.userID {
			var offlineMsg model.OfflineMessage
			if err := json.Unmarshal(message.Value, &offlineMsg); err != nil {
				global.GVA_LOG.Errorf("解析离线消息失败: %v", err)
				continue
			}
			h.messages = append(h.messages, &offlineMsg)
			session.MarkMessage(message, "")
		}
	}
	return nil
}

// MarkMessageAsRead 标记消息为已读
func (s *KafkaMessageStore) MarkMessageAsRead(messageID string) error {
	// 发送一个标记消息到Kafka
	markMsg := struct {
		Type      string `json:"type"`
		MessageID string `json:"message_id"`
		Action    string `json:"action"`
	}{
		Type:      "mark_read",
		MessageID: messageID,
		Action:    "read",
	}

	data, err := json.Marshal(markMsg)
	if err != nil {
		return err
	}

	return kafka.SendMessage(s.topic+".marks", messageID, data)
}

// DeleteMessage 删除消息
func (s *KafkaMessageStore) DeleteMessage(messageID string) error {
	// 发送一个删除消息到Kafka
	deleteMsg := struct {
		Type      string `json:"type"`
		MessageID string `json:"message_id"`
		Action    string `json:"action"`
	}{
		Type:      "mark_delete",
		MessageID: messageID,
		Action:    "delete",
	}

	data, err := json.Marshal(deleteMsg)
	if err != nil {
		return err
	}

	return kafka.SendMessage(s.topic+".marks", messageID, data)
}
