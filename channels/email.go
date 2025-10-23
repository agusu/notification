package channels

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/mail"
	"notification/models/channel"
	"os"
	"sync"
)

// ValidEmailMeta represents the required metadata for email notifications
type ValidEmailMeta struct {
	To       string `json:"to" example:"user@example.com"`
	Subject  string `json:"subject,omitempty" example:"Welcome to our platform"`
	Template string `json:"template,omitempty" example:"titled"`
}

type EmailChannel struct {
	templates map[string]*template.Template
	once      sync.Once
}

func (c *EmailChannel) initTemplates() {
	c.once.Do(func() {
		c.templates = make(map[string]*template.Template)
		c.templates["titled"] = template.Must(template.ParseFiles("email/titled.html.tmpl"))
		c.templates["plain"] = template.Must(template.ParseFiles("email/plain.txt.tmpl"))
	})
}

func (c *EmailChannel) getTemplate(templateName string) *template.Template {
	c.initTemplates()
	name := templateName
	if name == "" {
		name = "plain"
	}
	tmpl, ok := c.templates[name]
	if !ok {
		tmpl = c.templates["plain"]
	}
	return tmpl
}

func (c *EmailChannel) Name() string {
	return "email"
}

func (c *EmailChannel) Validate(meta map[string]string) error {
	to, ok := meta["to"]
	if !ok || to == "" {
		return fmt.Errorf("to field with valid email is required")
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func (c *EmailChannel) Send(ctx context.Context, msg channel.Message) error {
	tmpl := c.getTemplate(msg.Meta["template"])
	var body bytes.Buffer
	if err := tmpl.Execute(&body, msg); err != nil {
		return err
	}

	from := os.Getenv("EMAIL_FROM")
	to := msg.Meta["to"]
	subject := msg.Meta["subject"]

	return c.sender(ctx, from, to, subject, body.String())
}

func (c *EmailChannel) Prepare(ctx context.Context, msg *channel.Message) error {
	return nil
}

// separated from Send for testing purposes
func (c *EmailChannel) sender(ctx context.Context, from, to, subject, body string) error {
	fmt.Println(from, to, subject, body) // Implement actual email sending with external library or with net/smtp
	return nil
}
