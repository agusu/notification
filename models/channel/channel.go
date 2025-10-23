package channel

import (
	"context"
)

type Message struct {
	Title   string
	Content string
	Meta    map[string]string
}
type Channel interface {
	Name() string
	Send(ctx context.Context, msg Message) error
	Validate(meta map[string]string) error
	Prepare(ctx context.Context, msg *Message) error
}
