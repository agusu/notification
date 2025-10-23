package channels

import (
	"context"
	"fmt"
	"notification/models/channel"
	"regexp"
)

type SMSChannel struct{}

// ValidSMSMeta represents the required metadata for SMS notifications
type ValidSMSMeta struct {
	Phone    string `json:"phone" example:"+1234567890"`
	SendDate string `json:"send_date" example:"2024-10-21"`
}

func (c *SMSChannel) Name() string {
	return "sms"
}

func (c *SMSChannel) Send(ctx context.Context, msg channel.Message) error {
	return nil
}

func (c *SMSChannel) Validate(meta map[string]string) error {
	if phone, ok := meta["phone"]; !ok || !regexp.MustCompile(`^\+[1-9]\d{1,14}$`).MatchString(phone) {
		return fmt.Errorf("phone field with valid phone number is required")
	}
	if sendDate, ok := meta["send_date"]; !ok || !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(sendDate) {
		return fmt.Errorf("send_date field with valid date is required")
	}
	return nil
}

func (c *SMSChannel) Prepare(ctx context.Context, msg *channel.Message) error {
	if len(msg.Content) > 160 {
		msg.Content = msg.Content[:160]
	}
	return nil
}
