package channels

import (
	"context"
	"encoding/json"
	"io"
	"notification/models/channel"
	"os"
	"strings"
	"testing"
)

func TestPushValidate_OK(t *testing.T) {
	c := &PushChannel{}
	meta := map[string]string{
		"token": "valid_device_token_12345",
	}
	if err := c.Validate(meta); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestPushValidate_MissingToken(t *testing.T) {
	c := &PushChannel{}
	meta := map[string]string{}
	if err := c.Validate(meta); err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestPushValidate_EmptyToken(t *testing.T) {
	c := &PushChannel{}
	meta := map[string]string{"token": ""}
	if err := c.Validate(meta); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestPushValidate_TokenTooShort(t *testing.T) {
	c := &PushChannel{}
	meta := map[string]string{"token": "short"}
	if err := c.Validate(meta); err == nil {
		t.Fatal("expected error for token too short (< 10 chars)")
	}
}

func TestPushValidate_TokenTooLong(t *testing.T) {
	c := &PushChannel{}
	meta := map[string]string{"token": strings.Repeat("a", 5000)}
	if err := c.Validate(meta); err == nil {
		t.Fatal("expected error for token too long (> 4096 chars)")
	}
}

func TestPushSend_OK(t *testing.T) {
	c := &PushChannel{}
	msg := channel.Message{
		Title:   "Test Notification",
		Content: "This is a test push",
		Meta: map[string]string{
			"token": "device_token_abc123",
		},
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	if err := c.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	_ = w.Close()
	out, _ := io.ReadAll(r)
	output := string(out)

	if !strings.Contains(output, "Test Notification") {
		t.Fatalf("expected output to contain title, got: %q", output)
	}
	if !strings.Contains(output, "This is a test push") {
		t.Fatalf("expected output to contain content, got: %q", output)
	}
	if !strings.Contains(output, "device_token_abc123") {
		t.Fatalf("expected output to contain token, got: %q", output)
	}
}

func TestPushSend_WithData(t *testing.T) {
	c := &PushChannel{}
	dataJSON := `{"message_id":"123","type":"alert"}`
	msg := channel.Message{
		Title:   "Alert",
		Content: "New alert received",
		Meta: map[string]string{
			"token": "device_token_xyz",
			"data":  dataJSON,
		},
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	if err := c.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	_ = w.Close()
	out, _ := io.ReadAll(r)
	output := string(out)

	var payload pushPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if payload.Title != "Alert" {
		t.Fatalf("expected title 'Alert', got %q", payload.Title)
	}
	if payload.Body != "New alert received" {
		t.Fatalf("expected body 'New alert received', got %q", payload.Body)
	}
	if payload.Token != "device_token_xyz" {
		t.Fatalf("expected token 'device_token_xyz', got %q", payload.Token)
	}
	if payload.Data["message_id"] != "123" {
		t.Fatalf("expected data.message_id '123', got %q", payload.Data["message_id"])
	}
	if payload.Data["type"] != "alert" {
		t.Fatalf("expected data.type 'alert', got %q", payload.Data["type"])
	}
}

func TestPushSend_InvalidDataJSON(t *testing.T) {
	c := &PushChannel{}
	msg := channel.Message{
		Title:   "Test",
		Content: "Test content",
		Meta: map[string]string{
			"token": "device_token_test",
			"data":  "{invalid json",
		},
	}

	if err := c.Send(context.Background(), msg); err == nil {
		t.Fatal("expected error for invalid JSON in data")
	}
}

func TestPushSend_EmptyData(t *testing.T) {
	c := &PushChannel{}
	msg := channel.Message{
		Title:   "Test",
		Content: "Test content",
		Meta: map[string]string{
			"token": "device_token_test",
			"data":  "",
		},
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	if err := c.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	_ = w.Close()
	out, _ := io.ReadAll(r)
	output := string(out)

	var payload pushPayload
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if len(payload.Data) != 0 {
		t.Fatalf("expected empty data map, got %v", payload.Data)
	}
}

func TestPushName(t *testing.T) {
	c := &PushChannel{}
	if c.Name() != "push" {
		t.Fatalf("expected name 'push', got %q", c.Name())
	}
}
