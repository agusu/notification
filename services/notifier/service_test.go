package notifier

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"notification/models"
	"notification/models/channel"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeChannel struct {
	name        string
	validateErr error
	sendErr     error
	prepareErr  error
}

func (f *fakeChannel) Name() string                                            { return f.name }
func (f *fakeChannel) Validate(meta map[string]string) error                   { return f.validateErr }
func (f *fakeChannel) Send(ctx context.Context, msg channel.Message) error     { return f.sendErr }
func (f *fakeChannel) Prepare(ctx context.Context, msg *channel.Message) error { return f.prepareErr }

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
	db, err := gorm.Open(dsn, &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Notification{}, &models.Outbox{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestCreateAndEnqueue_OK(t *testing.T) {
	db := newTestDB(t)
	svc := NewNotifierService(db, map[string]channel.Channel{
		"email": &fakeChannel{name: "email"},
	})
	ctx := context.Background()

	req := NotificationRequest{Title: "t", Content: "c", ChannelName: "email", Meta: map[string]string{"k": "v"}}
	if err := svc.CreateAndEnqueue(ctx, req); err != nil {
		t.Fatalf("CreateAndEnqueue: %v", err)
	}

	var n models.Notification
	if err := db.First(&n).Error; err != nil {
		t.Fatalf("find notification: %v", err)
	}
	if n.Title != "t" || n.ChannelName != "email" {
		t.Fatalf("unexpected notification: %+v", n)
	}

	var o models.Outbox
	if err := db.First(&o).Error; err != nil {
		t.Fatalf("find outbox: %v", err)
	}
	if o.Status != models.PENDING {
		t.Fatalf("status: %v", o.Status)
	}
}

func TestCreateAndEnqueue_InvalidMeta(t *testing.T) {
	db := newTestDB(t)
	svc := NewNotifierService(db, map[string]channel.Channel{
		"email": &fakeChannel{name: "email", validateErr: errors.New("bad meta")},
	})
	ctx := context.Background()
	req := NotificationRequest{Title: "t", Content: "c", ChannelName: "email", Meta: map[string]string{"k": "v"}}
	if err := svc.CreateAndEnqueue(ctx, req); err == nil {
		t.Fatalf("expected error, got nil")
	}
	var count int64
	db.Model(&models.Outbox{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected no outbox rows, got %d", count)
	}
}

func TestDispatchOutbox_OK(t *testing.T) {
	db := newTestDB(t)
	svc := NewNotifierService(db, map[string]channel.Channel{
		"email": &fakeChannel{name: "email"},
	})
	// seed
	n := models.Notification{Title: "t", Content: "c", ChannelName: "email"}
	if err := db.Create(&n).Error; err != nil {
		t.Fatalf("seed notification: %v", err)
	}
	o := models.Outbox{NotificationID: n.ID, ChannelName: "email", PayloadJson: `{"title":"t","content":"c","meta":{}}`, Status: models.PROCESSING, NextAttemptAt: time.Now()}
	if err := db.Create(&o).Error; err != nil {
		t.Fatalf("seed outbox: %v", err)
	}

	if err := svc.DispatchOutbox(context.Background(), o); err != nil {
		t.Fatalf("DispatchOutbox: %v", err)
	}
}
