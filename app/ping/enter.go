package ping

import (
	"campus2/app/ping/controller"

	"github.com/gin-gonic/gin"
)

type PingApp struct {
	pingController *controller.PingController
}

func NewPingApp() *PingApp {
	return &PingApp{
		pingController: controller.NewPingController(),
	}
}

func (a *PingApp) InitPingRouter(private *gin.RouterGroup, public *gin.RouterGroup) {
	privateGroup := private.Group("ping")
	{
		privateGroup.GET("private", a.pingController.PrivatePing)
	}
	publicGroup := public.Group("ping")
	{
		publicGroup.GET("", a.pingController.PublicPing)
	}
}
