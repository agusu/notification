package notifier

import (
	"context"
	"log"
	"time"

	"notification/models"

	"gorm.io/gorm"
)

type Worker struct {
	db       *gorm.DB
	svc      *NotifierService
	interval time.Duration
	parallel int
}

func NewWorker(db *gorm.DB, svc *NotifierService, interval time.Duration, parallel int) *Worker {
	return &Worker{db: db, svc: svc, interval: interval, parallel: parallel}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jobs, err := w.claimBatch(ctx, w.parallel)
			if err != nil {
				log.Printf("Error fetching pending notifications: %v", err)
				continue
			}
			for _, job := range jobs {
				w.process(ctx, job)
			}
		}
	}
}

func (w *Worker) claimBatch(ctx context.Context, limit int) ([]models.Outbox, error) {
	tx := w.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var ids []int
	if err := tx.
		Model(&models.Outbox{}).
		Where("status = ? AND next_attempt_at <= ?", models.PENDING, time.Now()).
		Order("next_attempt_at ASC").
		Limit(limit).
		Pluck("id", &ids).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(ids) == 0 {
		tx.Rollback()
		return nil, nil
	}

	res := tx.Model(&models.Outbox{}).
		Where("id IN ? AND status = ?", ids, models.PENDING).
		Updates(map[string]any{"status": models.PROCESSING, "updated_at": time.Now()})
	if res.Error != nil {
		tx.Rollback()
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return nil, nil
	}

	var jobs []models.Outbox
	if err := tx.Where("id IN ?", ids).Find(&jobs).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return jobs, tx.Commit().Error
}

func (w *Worker) process(ctx context.Context, outbox models.Outbox) error {
	if outbox.Status != models.PROCESSING {
		return nil
	}
	return w.svc.DispatchOutbox(ctx, outbox)
}
