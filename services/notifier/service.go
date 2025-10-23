package notifier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"notification/models"
	"notification/models/channel"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidChannel       = errors.New("invalid channel name")
	ErrInvalidMetadata      = errors.New("invalid metadata for channel")
	ErrNotificationExists   = errors.New("notification already exists")
	ErrNotificationNotFound = errors.New("notification not found")
)

type NotificationRequest struct {
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	ChannelName string            `json:"channel_name"`
	Meta        map[string]string `json:"meta"`
	UserID      uint              `json:"user_id"`
}

type NotifierService struct {
	db          *gorm.DB
	channelList map[string]channel.Channel
}

func NewNotifierService(db *gorm.DB, channelList map[string]channel.Channel) *NotifierService {
	return &NotifierService{db: db, channelList: channelList}
}

func generateIdempotencyKey(notificationRequest NotificationRequest) (string, error) {
	// ensures only one notification is created for the same request
	payload := struct {
		UserID      uint              `json:"user_id"`
		ChannelName string            `json:"channel_name"`
		Title       string            `json:"title"`
		Content     string            `json:"content"`
		Meta        map[string]string `json:"meta"`
	}{notificationRequest.UserID, notificationRequest.ChannelName, notificationRequest.Title, notificationRequest.Content, notificationRequest.Meta}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}

func (s *NotifierService) CreateAndEnqueue(ctx context.Context, notificationRequest NotificationRequest) error {
	idempotencyKey, err := generateIdempotencyKey(notificationRequest)
	if err != nil {
		return err
	}
	notification := models.Notification{
		Title:          notificationRequest.Title,
		Content:        notificationRequest.Content,
		ChannelName:    notificationRequest.ChannelName,
		IdempotencyKey: idempotencyKey,
		UserID:         notificationRequest.UserID,
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Validar canal antes de crear la notificación
		channel, ok := s.channelList[notificationRequest.ChannelName]
		if !ok {
			return fmt.Errorf("%w: %s", ErrInvalidChannel, notificationRequest.ChannelName)
		}

		// Validar metadata del canal
		if err := channel.Validate(notificationRequest.Meta); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMetadata, err)
		}

		// Crear notificación
		err := tx.Create(&notification).Error
		if err != nil {
			// Si es por clave duplicada (idempotency), es un error de negocio
			return err
		}

		payloadBody := struct {
			Title   string            `json:"title"`
			Content string            `json:"content"`
			Meta    map[string]string `json:"meta"`
		}{
			Title:   notification.Title,
			Content: notification.Content,
			Meta:    notificationRequest.Meta,
		}

		payload, err := json.Marshal(payloadBody)
		if err != nil {
			return err
		}

		outbox := models.Outbox{
			NotificationID: notification.ID,
			ChannelName:    notificationRequest.ChannelName,
			PayloadJson:    string(payload),
			Status:         models.PENDING,
			Attempts:       0,
			LastError:      "",
			NextAttemptAt:  time.Now(),
			MaxAttempts:    3,
		}

		if err := tx.Create(&outbox).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *NotifierService) DispatchOutbox(ctx context.Context, outbox models.Outbox) error {
	var message channel.Message
	err := json.Unmarshal([]byte(outbox.PayloadJson), &message)
	if err != nil {
		return fmt.Errorf("internal error")
	}

	channel, ok := s.channelList[outbox.ChannelName]
	if !ok {
		return fmt.Errorf("channel %s not found", outbox.ChannelName)
	}

	err = channel.Send(ctx, message)
	if err != nil {
		return err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Outbox{}).
			Where("id=? AND status=?", outbox.ID, models.PROCESSING).
			Update("status", models.SENT).
			Update("updated_at", time.Now())
		if res.Error != nil {
			err = tx.Model(&models.Outbox{}).
				Where("id = ?", outbox.ID).
				Update("attempts", outbox.Attempts+1).
				Update("next_attempt_at", time.Now().Add(time.Minute)).
				Update("last_error", res.Error.Error()).
				Error
			return res.Error
		}
		return nil
	})
	return err
}

func (s *NotifierService) GetNotification(ctx context.Context, id int) (*models.Notification, error) {
	var n models.Notification
	if err := s.db.WithContext(ctx).First(&n, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, err
	}
	return &n, nil
}

func (s *NotifierService) ListNotifications(ctx context.Context, userID int, limit, offset int) ([]models.Notification, error) {
	var list []models.Notification
	q := s.db.WithContext(ctx).Order("created_at DESC")
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *NotifierService) UpdateNotification(ctx context.Context, id int, patch map[string]any) error {
	allowed := map[string]bool{"title": true, "content": true}
	filtered := map[string]any{}
	for k, v := range patch {
		if allowed[k] {
			filtered[k] = v
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	result := s.db.WithContext(ctx).Model(&models.Notification{}).Where("id = ?", id).Updates(filtered)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

func (s *NotifierService) DeleteNotification(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}
