package model

import (
	"campus2/pkg/global"
	"time"

	"gorm.io/gorm"
)

// PongModel 用于记录ping日志
type PongModel struct {
	gorm.Model
	Message   string    `gorm:"size:255;not null"`
	ClientIP  string    `gorm:"size:64;not null"`
	PingTime  time.Time `gorm:"not null"`
	IsPrivate bool      `gorm:"not null;default:false"`
}

// CreatePingLog 创建ping日志
func (p *PongModel) CreatePingLog() error {
	return global.GVA_DB.Create(p).Error
}
