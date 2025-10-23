package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"notification/models/channel"
)

type PushChannel struct {
}

// ValidPushMeta represents the required metadata for push notifications
type ValidPushMeta struct {
	Token    string            `json:"token" example:"device_token_xyz123"`
	Platform string            `json:"platform" example:"android"`
	Data     map[string]string `json:"data,omitempty" swaggertype:"object,string"`
	Options  map[string]string `json:"options,omitempty" swaggertype:"object,string"`
}

type pushPayload struct {
	Token string            `json:"token"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"`
}

func (c *PushChannel) Name() string {
	return "push"
}

func (c *PushChannel) Send(ctx context.Context, msg channel.Message) error {
	data := map[string]string{}
	if s := msg.Meta["data"]; s != "" {
		if err := json.Unmarshal([]byte(s), &data); err != nil {
			return fmt.Errorf("invalid data json: %w", err)
		}
	}
	payload := pushPayload{
		Title: msg.Title,
		Body:  msg.Content,
		Data:  data,
		Token: msg.Meta["token"],
	}

	b, _ := json.Marshal(payload)
	fmt.Println(string(b)) // Replace with actual push notification sending logic
	return nil
}

func (c *PushChannel) Validate(meta map[string]string) error {
	token := meta["token"]
	if token == "" || len(token) < 10 || len(token) > 4096 {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (c *PushChannel) Prepare(ctx context.Context, msg *channel.Message) error {
	return nil
}
