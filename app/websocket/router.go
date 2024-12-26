package websocket

import (
	"github.com/gin-gonic/gin"
)

type WebSocketApp struct {
	handler *Handler
	manager *Manager
}

func NewWebSocketApp() *WebSocketApp {
	manager := NewManager()
	go manager.Start() // 启动WebSocket管理器

	return &WebSocketApp{
		handler: NewHandler(manager),
		manager: manager,
	}
}

func (app *WebSocketApp) InitWebSocketRouter(r *gin.Engine) {
	ws := r.Group("/ws")
	{
		ws.GET("", app.handler.HandleWebSocket)
	}
}
