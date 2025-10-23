package models

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
}
