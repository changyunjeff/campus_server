package model

import (
	"time"
)

// 消息类型常量
const (
	MessageTypeChat    = "chat"    // 聊天消息
	MessageTypeLike    = "like"    // 点赞通知
	MessageTypeCollect = "collect" // 收藏通知
	MessageTypeComment = "comment" // 评论通知
	MessageTypeMention = "mention" // @通知
	MessageTypeSystem  = "system"  // 系统消息
)

// Message 消息结构
type Message struct {
	Type      string       `json:"type"`      // 消息类型
	Content   interface{}  `json:"content"`   // 消息内容
	From      string       `json:"from"`      // 发送者ID
	To        string       `json:"to"`        // 接收者ID
	CreatedAt time.Time    `json:"createdAt"` // 创建时间
	Extra     MessageExtra `json:"extra"`     // 额外信息
}

// MessageExtra 消息额外信息
type MessageExtra struct {
	PostID     string `json:"postId,omitempty"`     // 动态ID
	CommentID  string `json:"commentId,omitempty"`  // 评论ID
	ActionType string `json:"actionType,omitempty"` // 动作类型(like/unlike/collect/uncollect等)
	URL        string `json:"url,omitempty"`        // 相关链接
}

// OfflineMessage 离线消息模型
type OfflineMessage struct {
	ID        string       `json:"id"`        // 消息ID
	Type      string       `json:"type"`      // 消息类型
	Content   interface{}  `json:"content"`   // 消息内容
	From      string       `json:"from"`      // 发送者
	To        string       `json:"to"`        // 接收者
	Timestamp time.Time    `json:"timestamp"` // 发送时间
	Status    int          `json:"status"`    // 消息状态(0:未读,1:已读)
	Extra     MessageExtra `json:"extra"`     // 额外信息
}

// MessageStore 消息存储接口
type MessageStore interface {
	// 存储离线消息
	StoreMessage(msg *OfflineMessage) error
	// 获取用户的离线消息
	GetOfflineMessages(userID string) ([]*OfflineMessage, error)
	// 标记消息为已读
	MarkMessageAsRead(messageID string) error
	// 删除消息
	DeleteMessage(messageID string) error
}
