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
	Phone   string `json:"phone" example:"+1234567890"`
	Carrier string `json:"carrier" example:"verizon"`
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
	if carrier, ok := meta["carrier"]; !ok || carrier == "" {
		return fmt.Errorf("carrier field is required")
	}
	return nil
}

func (c *SMSChannel) Prepare(ctx context.Context, msg *channel.Message) error {
	if len(msg.Content) > 160 {
		msg.Content = msg.Content[:160]
	}
	return nil
}
