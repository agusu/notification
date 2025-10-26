package channels

import (
	"context"
	"notification/models/channel"
	"strings"
	"testing"
)

func TestSMSValidate_OK(t *testing.T) {
	c := &SMSChannel{}
	meta := map[string]string{
		"phone":   "+1234567890",
		"carrier": "verizon",
	}
	if err := c.Validate(meta); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}

func TestSMSValidate_MissingPhone(t *testing.T) {
	c := &SMSChannel{}
	meta := map[string]string{"carrier": "verizon"}
	if err := c.Validate(meta); err == nil {
		t.Fatal("expected error for missing phone")
	}
}

func TestSMSValidate_InvalidPhoneFormat(t *testing.T) {
	c := &SMSChannel{}
	tests := []struct {
		name  string
		phone string
	}{
		{"no plus", "1234567890"},
		{"starts with zero", "+0123456789"},
		{"too short", "+1"},
		{"letters", "+123abc"},
		{"spaces", "+1 234 567 890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := map[string]string{
				"phone":   tt.phone,
				"carrier": "verizon",
			}
			if err := c.Validate(meta); err == nil {
				t.Fatalf("expected error for invalid phone: %s", tt.phone)
			}
		})
	}
}

func TestSMSValidate_MissingCarrier(t *testing.T) {
	c := &SMSChannel{}
	meta := map[string]string{"phone": "+1234567890"}
	if err := c.Validate(meta); err == nil {
		t.Fatal("expected error for missing carrier")
	}
}

func TestSMSValidate_EmptyCarrier(t *testing.T) {
	c := &SMSChannel{}
	meta := map[string]string{
		"phone":   "+1234567890",
		"carrier": "",
	}
	if err := c.Validate(meta); err == nil {
		t.Fatal("expected error for empty carrier")
	}
}

func TestSMSSend_OK(t *testing.T) {
	c := &SMSChannel{}
	msg := channel.Message{
		Title:   "Test",
		Content: "Hello SMS",
		Meta: map[string]string{
			"phone":   "+1234567890",
			"carrier": "att",
		},
	}
	if err := c.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

func TestSMSPrepare_TruncatesLongContent(t *testing.T) {
	c := &SMSChannel{}
	longContent := strings.Repeat("a", 200)
	msg := channel.Message{
		Title:   "Test",
		Content: longContent,
		Meta: map[string]string{
			"phone":   "+1234567890",
			"carrier": "verizon",
		},
	}

	if err := c.Prepare(context.Background(), &msg); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if len(msg.Content) != 160 {
		t.Fatalf("expected content length 160, got %d", len(msg.Content))
	}
}

func TestSMSPrepare_ShortContentUnchanged(t *testing.T) {
	c := &SMSChannel{}
	shortContent := "Short message"
	msg := channel.Message{
		Title:   "Test",
		Content: shortContent,
		Meta: map[string]string{
			"phone":   "+1234567890",
			"carrier": "tmobile",
		},
	}

	if err := c.Prepare(context.Background(), &msg); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if msg.Content != shortContent {
		t.Fatalf("expected content %q, got %q", shortContent, msg.Content)
	}
}

func TestSMSName(t *testing.T) {
	c := &SMSChannel{}
	if c.Name() != "sms" {
		t.Fatalf("expected name 'sms', got %q", c.Name())
	}
}
