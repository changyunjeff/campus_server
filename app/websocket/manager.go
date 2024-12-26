package websocket

import (
	"campus2/pkg/global"
	"context"
	"encoding/json"
	"sync"
	"time"

	"campus2/app/websocket/model"
	"campus2/app/websocket/store"

	"github.com/gorilla/websocket"
)

// Client WebSocket客户端
type Client struct {
	ID       string
	UserID   string
	Socket   *websocket.Conn
	Send     chan []byte
	Manager  *Manager
	LastPing time.Time
}

// Manager WebSocket管理器
type Manager struct {
	clients    sync.Map     // 本地连接的客户端 map[string]*Client
	broadcast  chan []byte  // 广播消息通道
	register   chan *Client // 注册通道
	unregister chan *Client // 注销通道
	// 根据配置决定是否初始化存储
	redisStore *store.RedisMessageStore `json:"-"`
	kafkaStore *store.KafkaMessageStore `json:"-"`
}

// ConnInfo 连接信息
type ConnInfo struct {
	UserID   string `json:"user_id"`
	ServerID string `json:"server_id"` // 服务器标识
	LastPing int64  `json:"last_ping"`
}

const (
	// Redis key 前缀
	connMapKey    = "ws:conn:map"     // Hash表存储用户连接信息
	serverConnKey = "ws:server:conns" // Set存储服务器的在线连接
)

// NewManager 创建WebSocket管理器
func NewManager() *Manager {
	m := &Manager{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	// 根据配置初始化存储
	if global.GVA_CONFIG.System.UseRedis {
		m.redisStore = store.NewRedisMessageStore(global.GVA_CONFIG.Redis.GetDuration())
	}
	if global.GVA_CONFIG.System.UseKafka {
		m.kafkaStore = store.NewKafkaMessageStore(global.GVA_CONFIG.Kafka.Topic)
	}

	return m
}

// Start 启动WebSocket管理器
func (m *Manager) Start() {
	global.GVA_LOG.Info("WebSocket管理器开始运行")
	for {
		select {
		case client := <-m.register:
			global.GVA_LOG.Infof("注册新的WebSocket客户端: %s, 用户ID: %s", client.ID, client.UserID)
			// 注册客户端到本地
			m.clients.Store(client.ID, client)
			// 只在启用Redis时更新连接信息
			if m.redisStore != nil {
				m.updateConnInfo(client)
			}
			global.GVA_LOG.Infof("客户端注册完成: %s", client.ID)

		case client := <-m.unregister:
			global.GVA_LOG.Infof("注销WebSocket客户端: %s, 用户ID: %s", client.ID, client.UserID)
			if _, ok := m.clients.Load(client.ID); ok {
				m.clients.Delete(client.ID)
				close(client.Send)
				// 只在启用Redis时移除连接信息
				if m.redisStore != nil {
					m.removeConnInfo(client)
				}
				global.GVA_LOG.Infof("客户端注销完成: %s", client.ID)
			}

		case message := <-m.broadcast:
			global.GVA_LOG.Info("收到广播消息，准备向所有在线用户推送")
			// 直接从本地clients广播
			m.clients.Range(func(key, value interface{}) bool {
				client := value.(*Client)
				select {
				case client.Send <- message:
					global.GVA_LOG.Infof("广播消息已发送给用户: %s", client.UserID)
				default:
					close(client.Send)
					m.clients.Delete(client.ID)
				}
				return true
			})
		}
	}
}

// updateConnInfo 更新Redis中的连接信息
func (m *Manager) updateConnInfo(client *Client) {
	ctx := context.Background()
	connInfo := ConnInfo{
		UserID:   client.UserID,
		ServerID: global.GVA_CONFIG.System.ServerID,
		LastPing: time.Now().Unix(),
	}

	data, err := json.Marshal(connInfo)
	if err != nil {
		global.GVA_LOG.Errorf("序列化连接信息失败: %v", err)
		return
	}

	global.GVA_LOG.Infof("更新用户 %s 的连接信息到Redis", client.UserID)
	// 使用Redis Hash存储用户连接信息
	pipe := global.GVA_REDIS.Pipeline()
	pipe.HSet(ctx, connMapKey, client.UserID, data)
	pipe.Expire(ctx, connMapKey, global.GVA_CONFIG.Redis.GetDuration())
	pipe.HSet(ctx, "online:last_seen", client.UserID, time.Now().Unix())
	pipe.Expire(ctx, "online:last_seen", global.GVA_CONFIG.Redis.GetDuration())
	_, err = pipe.Exec(ctx)
	if err != nil {
		global.GVA_LOG.Errorf("更新连接信息到Redis失败: %v", err)
	} else {
		global.GVA_LOG.Infof("成功更新用户 %s 的连接信息", client.UserID)
	}
}

// removeConnInfo 从Redis中移除连接信息
func (m *Manager) removeConnInfo(client *Client) {
	ctx := context.Background()
	global.GVA_LOG.Infof("从Redis中移除用户 %s 的连接信息", client.UserID)
	err := global.GVA_REDIS.HDel(ctx, connMapKey, client.UserID).Err()
	if err != nil {
		global.GVA_LOG.Errorf("从Redis移除连接信息失败: %v", err)
	} else {
		global.GVA_LOG.Infof("成功移除用户 %s 的连接信息", client.UserID)
	}
}

// SendToUser 发送消息给指定用户
func (m *Manager) SendToUser(userID string, message []byte) error {
	global.GVA_LOG.Infof("准备向用户 %s 发送消息", userID)

	var messageSent bool
	// 在本地查找用户
	m.clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		if client.UserID == userID {
			select {
			case client.Send <- message:
				global.GVA_LOG.Infof("成功向用户 %s 发送消息", userID)
				messageSent = true
				return false
			default:
				global.GVA_LOG.Warnf("向用户 %s 发送消息失败，清理连接", userID)
				close(client.Send)
				m.clients.Delete(client.ID)
			}
		}
		return true
	})

	if !messageSent {
		// 如果启用了Redis/Kafka，则存储离线消息
		if m.redisStore != nil || m.kafkaStore != nil {
			var msg model.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				return err
			}

			offlineMsg := &model.OfflineMessage{
				ID:        time.Now().Format("20060102150405") + ":" + msg.From,
				Type:      msg.Type,
				Content:   msg.Content,
				From:      msg.From,
				To:        msg.To,
				Timestamp: time.Now(),
				Status:    0,
			}

			if m.redisStore != nil {
				if err := m.redisStore.StoreMessage(offlineMsg); err != nil {
					global.GVA_LOG.Errorf("存储离线消息到Redis失败: %v", err)
				}
			}

			if m.kafkaStore != nil {
				if err := m.kafkaStore.StoreMessage(offlineMsg); err != nil {
					global.GVA_LOG.Warnf("备份离线消息到Kafka失败: %v", err)
				}
			}
		} else {
			global.GVA_LOG.Warnf("用户 %s 不在线，且未启用离线消息存储", userID)
		}
	}

	return nil
}

// GetOnlineUsers 获取在线用户列表
func (m *Manager) GetOnlineUsers() ([]string, error) {
	ctx := context.Background()
	global.GVA_LOG.Info("获取在线用户列表")
	users, err := global.GVA_REDIS.HKeys(ctx, connMapKey).Result()
	if err != nil {
		global.GVA_LOG.Errorf("获取在线用户列表失败: %v", err)
		return nil, err
	}
	global.GVA_LOG.Infof("当前在线用户数量: %d", len(users))
	return users, nil
}

// IsUserOnline 检查用户是否在线
func (m *Manager) IsUserOnline(userID string) (bool, error) {
	ctx := context.Background()
	global.GVA_LOG.Infof("检查用户 %s 是否在线", userID)
	exists, err := global.GVA_REDIS.HExists(ctx, connMapKey, userID).Result()
	if err != nil {
		global.GVA_LOG.Errorf("检查用户 %s 在线状态失败: %v", userID, err)
		return false, err
	}
	global.GVA_LOG.Infof("用户 %s 在线状态: %v", userID, exists)
	return exists, nil
}
