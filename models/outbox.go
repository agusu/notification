package models

import "time"

type Status string

const (
	PENDING    Status = "PENDING"
	PROCESSING Status = "PROCESSING"
	SENT       Status = "SENT"
	FAILED     Status = "FAILED"
)

type Outbox struct {
	ID             uint
	NotificationID uint
	ChannelName    string
	PayloadJson    string
	Status         Status `gorm:"index:idx_status_scheduled,priority:1"`
	Attempts       int
	LastError      string
	NextAttemptAt  time.Time
	ScheduledAt    time.Time `gorm:"index:idx_status_scheduled,priority:2"`
	MaxAttempts    int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
