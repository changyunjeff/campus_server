package service

import (
	"campus2/app/ping/model"
	"campus2/app/ping/vo"
	"time"
)

type PongService struct{}

func NewPongService() *PongService {
	return &PongService{}
}

// HandlePublicPing 处理公开ping
func (s *PongService) HandlePublicPing(clientIP, echo string) (*vo.Pong, error) {
	message := "pong"
	if echo != "" {
		message = echo
	}

	// 记录日志
	pingLog := &model.PongModel{
		Message:   message,
		ClientIP:  clientIP,
		PingTime:  time.Now(),
		IsPrivate: false,
	}
	if err := pingLog.CreatePingLog(); err != nil {
		return nil, err
	}

	return &vo.Pong{
		Message: message,
		Time:    time.Now().UnixNano() / 1e6,
	}, nil
}

// HandlePrivatePing 处理私有ping
func (s *PongService) HandlePrivatePing(clientIP, echo string) (*vo.Pong, error) {
	message := "private pong"
	if echo != "" {
		message = "private " + echo
	}

	// 记录日志
	pingLog := &model.PongModel{
		Message:   message,
		ClientIP:  clientIP,
		PingTime:  time.Now(),
		IsPrivate: true,
	}
	if err := pingLog.CreatePingLog(); err != nil {
		return nil, err
	}

	return &vo.Pong{
		Message: message,
		Time:    time.Now().UnixNano() / 1e6,
	}, nil
}
