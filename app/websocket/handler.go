package websocket

import (
	"campus2/app/websocket/model"
	"campus2/pkg/global"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// Handler WebSocket处理器
type Handler struct {
	manager *Manager
}

// NewHandler 创建WebSocket处理器
func NewHandler(manager *Manager) *Handler {
	return &Handler{manager: manager}
}

// HandleWebSocket 处理WebSocket连接
func (h *Handler) HandleWebSocket(c *gin.Context) {
	userID := c.Query("user_id") // 从查询参数获取用户ID
	if userID == "" {
		global.GVA_LOG.Error("用户ID为空，连接失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}
	global.GVA_LOG.Info("开始处理WebSocket连接，获取userID:", userID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.GVA_LOG.Errorf("协议升级失败: %v", err)
		return
	}
	global.GVA_LOG.Info("成功将http协议升级为ws协议")

	client := &Client{
		ID:       userID + ":" + time.Now().String(),
		UserID:   userID,
		Socket:   conn,
		Send:     make(chan []byte, 256),
		Manager:  h.manager,
		LastPing: time.Now(),
	}
	global.GVA_LOG.Infof("创建新的客户端: %s", client.ID)

	h.manager.register <- client
	global.GVA_LOG.Infof("向WebSocket管理器注册客户端:%v", client.ID)

	// 只有在启用存储时才获取离线消息
	if h.manager.redisStore != nil {
		messages, err := h.manager.redisStore.GetOfflineMessages(userID)
		if err != nil {
			global.GVA_LOG.Errorf("获取离线消息失败: %v", err)
		} else {
			global.GVA_LOG.Infof("获取到 %d 条离线消息", len(messages))
			for _, msg := range messages {
				data, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				client.Send <- data
			}
		}
	}

	// 启动读写goroutine
	global.GVA_LOG.Infof("启动客户端 %s 的读写协程", client.ID)
	go client.writePump()
	go client.readPump()
}

// writePump 处理向客户端写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(time.Second * time.Duration(global.GVA_CONFIG.WebSocket.HeartbeatTime))
	defer func() {
		global.GVA_LOG.Infof("客户端 %s 的写入协程结束", c.ID)
		ticker.Stop()
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				global.GVA_LOG.Infof("客户端 %s 的发送通道已关闭", c.ID)
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Socket.NextWriter(websocket.TextMessage)
			if err != nil {
				global.GVA_LOG.Errorf("客户端 %s 创建消息写入器失败: %v", c.ID, err)
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				global.GVA_LOG.Errorf("客户端 %s 关闭消息写入器失败: %v", c.ID, err)
				return
			}
			global.GVA_LOG.Infof("客户端 %s 成功发送消息", c.ID)
		case <-ticker.C:
			if err := c.Socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				global.GVA_LOG.Errorf("客户端 %s 发送心跳包失败: %v", c.ID, err)
				return
			}
			global.GVA_LOG.Debugf("客户端 %s 发送心跳包", c.ID)
		}
	}
}

// readPump 处理从客户端读取消息
func (c *Client) readPump() {
	defer func() {
		global.GVA_LOG.Infof("客户端 %s 的读取协程结束，准备注销", c.ID)
		c.Manager.unregister <- c
		c.Socket.Close()
	}()

	c.Socket.SetReadLimit(int64(global.GVA_CONFIG.WebSocket.ReadBufferSize))
	global.GVA_LOG.Infof("设置客户端 %s 的读取缓冲区大小为 %d", c.ID, global.GVA_CONFIG.WebSocket.ReadBufferSize)

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				global.GVA_LOG.Errorf("客户端 %s 读取错误: %v", c.ID, err)
			}
			break
		}

		// 处理收到的消息
		var msg model.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			global.GVA_LOG.Errorf("客户端 %s 解析消息失败: %v", c.ID, err)
			continue
		}

		msg.From = c.UserID // 设置发送者ID
		msg.CreatedAt = time.Now()

		global.GVA_LOG.Infof("收到客户端 %s 的消息: type=%s, from=%s, to=%s", c.ID, msg.Type, msg.From, msg.To)

		// 根据消息类型处理
		switch msg.Type {
		case model.MessageTypeChat:
			// 处理聊天消息
			if msg.To != "" {
				global.GVA_LOG.Infof("客户端 %s 发送私聊消息给用户 %s", c.ID, msg.To)
				data, _ := json.Marshal(msg)
				c.Manager.SendToUser(msg.To, data)
			} else {
				global.GVA_LOG.Info("客户端 %s 发送广播消息", c.ID)
				data, _ := json.Marshal(msg)
				c.Manager.broadcast <- data
			}

		case model.MessageTypeLike:
			// 处理点赞通知
			if msg.Extra.PostID != "" {
				global.GVA_LOG.Infof("用户 %s 点赞了动态 %s", c.UserID, msg.Extra.PostID)
				data, _ := json.Marshal(msg)
				c.Manager.SendToUser(msg.To, data)
			}

		case model.MessageTypeCollect:
			// 处理收藏通知
			if msg.Extra.PostID != "" {
				global.GVA_LOG.Infof("用户 %s 收藏了动态 %s", c.UserID, msg.Extra.PostID)
				data, _ := json.Marshal(msg)
				c.Manager.SendToUser(msg.To, data)
			}

		case model.MessageTypeComment:
			// 处理评论通知
			if msg.Extra.PostID != "" {
				global.GVA_LOG.Infof("用户 %s 评论了动态 %s", c.UserID, msg.Extra.PostID)
				data, _ := json.Marshal(msg)
				c.Manager.SendToUser(msg.To, data)
			}

		case model.MessageTypeMention:
			// 处理@通知
			if msg.To != "" {
				global.GVA_LOG.Infof("用户 %s 在动态/评论中@了用户 %s", c.UserID, msg.To)
				data, _ := json.Marshal(msg)
				c.Manager.SendToUser(msg.To, data)
			}

		case "ping":
			// 更新最后心跳时间
			c.LastPing = time.Now()
			global.GVA_LOG.Debugf("更新客户端 %s 的心跳时间: %v", c.ID, c.LastPing)
		}
	}
}
