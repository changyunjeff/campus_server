package store

import (
	"campus2/app/websocket/model"
	"campus2/pkg/global"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type RedisMessageStore struct {
	expiration time.Duration // 消息过期时间
}

func NewRedisMessageStore(expiration time.Duration) *RedisMessageStore {
	return &RedisMessageStore{
		expiration: expiration,
	}
}

func (s *RedisMessageStore) StoreMessage(msg *model.OfflineMessage) error {
	ctx := context.Background()
	key := fmt.Sprintf("offline:msg:%s", msg.To)

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 使用Redis List存储消息
	err = global.GVA_REDIS.LPush(ctx, key, data).Err()
	if err != nil {
		return err
	}

	// 设置过期时间
	return global.GVA_REDIS.Expire(ctx, key, s.expiration).Err()
}

func (s *RedisMessageStore) GetOfflineMessages(userID string) ([]*model.OfflineMessage, error) {
	ctx := context.Background()
	key := fmt.Sprintf("offline:msg:%s", userID)

	// 获取所有消息
	data, err := global.GVA_REDIS.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]*model.OfflineMessage, 0, len(data))
	for _, item := range data {
		var msg model.OfflineMessage
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}

	// 获取后删除消息
	global.GVA_REDIS.Del(ctx, key)

	return messages, nil
}
