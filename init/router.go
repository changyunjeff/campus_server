package init

import (
	"campus2/app/ping"
	"campus2/app/websocket"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.New()

	// 使用日志和恢复中间件
	Router.Use(gin.Recovery())
	if gin.Mode() == gin.DebugMode {
		Router.Use(gin.Logger())
	}

	private := Router.Group("")
	public := Router.Group("")

	// 注册 ping 路由
	ping.NewPingApp().InitPingRouter(private, public)

	// 注册WebSocket路由
	websocket.NewWebSocketApp().InitWebSocketRouter(Router)

	return Router
}
