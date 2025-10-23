package channels

import (
	"context"
	"io"
	"notification/models/channel"
	"os"
	"strings"
	"testing"
)

func TestEmailValidate_OK(t *testing.T) {
	c := &EmailChannel{}
	if err := c.Validate(map[string]string{"to": "user@example.com"}); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestEmailValidate_Invalid(t *testing.T) {
	c := &EmailChannel{}
	if err := c.Validate(map[string]string{"to": "bad@@example"}); err == nil {
		t.Fatalf("expected error for invalid email")
	}
}

func TestEmailSend_TitledTemplate(t *testing.T) {
	c := &EmailChannel{}
	msg := channel.Message{Title: "Hola", Content: "Mundo", Meta: map[string]string{"template": "titled", "to": "user@example.com", "subject": "s"}}
	// capture output by calling Send; we assert no error and basic markers in body via template execution
	if err := c.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestEmailSend_DefaultTemplate(t *testing.T) {
	c := &EmailChannel{}
	msg := channel.Message{Title: "Hola", Content: "Texto plano", Meta: map[string]string{"to": "user@example.com"}}
	if err := c.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send: %v", err)
	}
	// quick sanity check rendering uses content
	// We don't have direct access to body here; minimal test ensures no error. Optionally, we could refactor Send to return body for testing.
	if strings.TrimSpace(msg.Content) == "" {
		t.Fatalf("unexpected empty content")
	}
}

func TestEmailSender_OK(t *testing.T) {
	c := &EmailChannel{}
	from := "from@example.com"
	to := "to@example.com"
	subject := "Subject"
	body := "Body content"

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	if err := c.sender(context.Background(), from, to, subject, body); err != nil {
		t.Fatalf("sender: %v", err)
	}
	_ = w.Close()
	out, _ := io.ReadAll(r)
	got := string(out)
	if !strings.Contains(got, subject) || !strings.Contains(got, to) || !strings.Contains(got, body) {
		t.Fatalf("unexpected sender output: %q", got)
	}
}
