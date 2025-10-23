package models

import (
	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model
	UserID         uint `gorm:"not null;index"`
	Title          string
	Content        string
	ChannelName    string
	IdempotencyKey string
}
