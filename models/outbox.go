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
	Status         Status
	Attempts       int
	LastError      string
	NextAttemptAt  time.Time
	MaxAttempts    int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
