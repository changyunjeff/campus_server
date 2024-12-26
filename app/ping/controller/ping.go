package controller

import (
	"campus2/app/ping/dto"
	"campus2/app/ping/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PingController struct {
	pongService *service.PongService
}

func NewPingController() *PingController {
	return &PingController{
		pongService: service.NewPongService(),
	}
}

// PublicPing godoc
// @Summary 公开ping接口
// @Description 返回pong和当前时间戳
// @Tags 系统
// @Accept json
// @Produce json
// @Param echo query string false "回声参数"
// @Success 200 {object} vo.Pong
// @Router /api/v1/ping/public [get]
func (pc *PingController) PublicPing(c *gin.Context) {
	var req dto.PingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := pc.pongService.HandlePublicPing(c.ClientIP(), req.Echo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PrivatePing godoc
// @Summary 私有ping接口
// @Description 返回private pong和当前时间戳
// @Tags 系统
// @Accept json
// @Produce json
// @Param echo query string false "回声参数"
// @Success 200 {object} vo.Pong
// @Router /api/v1/ping/private [get]
func (pc *PingController) PrivatePing(c *gin.Context) {
	var req dto.PingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := pc.pongService.HandlePrivatePing(c.ClientIP(), req.Echo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
