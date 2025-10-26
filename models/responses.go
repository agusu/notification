package models

import "time"

// ErrorResponse represents an error response body
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}

// MessageResponse represents a success message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// TokenResponse represents a login response with JWT token
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// CreateNotificationRequest represents the request body for creating a notification
type CreateNotificationRequest struct {
	Title       string            `json:"title" example:"Welcome email"`
	Content     string            `json:"content" example:"Welcome to our platform!"`
	ChannelName string            `json:"channel_name" example:"email" enums:"email,sms,push"`
	Meta        map[string]string `json:"meta" swaggertype:"object,string"`
	ScheduledAt *string           `json:"scheduled_at,omitempty" example:"2025-10-27T10:00:00Z"`
}

// NotificationResponse represents a notification for API responses (without gorm.Model)
type NotificationResponse struct {
	ID             uint      `json:"id" example:"1"`
	CreatedAt      time.Time `json:"created_at" example:"2025-10-26T12:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" example:"2025-10-26T12:00:00Z"`
	UserID         uint      `json:"user_id" example:"123"`
	Title          string    `json:"title" example:"Welcome email"`
	Content        string    `json:"content" example:"Welcome to our platform!"`
	ChannelName    string    `json:"channel_name" example:"email"`
	IdempotencyKey string    `json:"idempotency_key" example:"a1b2c3d4e5f6"`
}
