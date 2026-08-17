package service

import (
	"context"
	"fmt"

	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

type RetryService struct {
	store *storage.Store
	clock Clock
}

func NewRetryService(store *storage.Store, clock Clock) *RetryService {
	return &RetryService{store: store, clock: clock}
}

func (r *RetryService) Queue(record domain.SyncRecord, message string) error {
	task := domain.RetryTask{ID: "retry:" + record.ID, RecordID: record.ID, Vendor: record.Vendor, MaxAttempts: 3}
	task.Schedule(r.clock.Now(), message)
	if err := r.store.SaveRetryTask(task); err != nil {
		return err
	}
	_, err := r.appendAudit(record.ID, domain.ActionRetryQueued, "retry queued: "+message)
	return err
}

func (r *RetryService) Due() ([]domain.RetryTask, error) {
	tasks, err := r.store.ListRetryTasks()
	if err != nil {
		return nil, err
	}
	due := make([]domain.RetryTask, 0, len(tasks))
	for _, task := range tasks {
		if task.IsDue(r.clock.Now()) {
			due = append(due, task)
		}
	}
	return due, nil
}

func (r *RetryService) Resolve(id string) error {
	tasks, err := r.store.ListRetryTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.ID == id {
			task.Resolve()
			return r.store.SaveRetryTask(task)
		}
	}
	return fmt.Errorf("retry task %q not found", id)
}

func (r *RetryService) Reschedule(id, message string) error {
	tasks, err := r.store.ListRetryTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.ID == id {
			task.Schedule(r.clock.Now(), message)
			return r.store.SaveRetryTask(task)
		}
	}
	return fmt.Errorf("retry task %q not found", id)
}

func (r *RetryService) Run(ctx context.Context, handler func(context.Context, domain.RetryTask) error) error {
	tasks, err := r.Due()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if runErr := handler(ctx, task); runErr != nil {
			if err := r.Reschedule(task.ID, runErr.Error()); err != nil {
				return err
			}
			continue
		}
		if err := r.Resolve(task.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *RetryService) appendAudit(recordID string, action domain.AuditAction, message string) (domain.AuditEvent, error) {
	event := domain.NewAuditEvent(fmt.Sprintf("retry-audit:%s:%d", recordID, r.clock.Now().UnixNano()), recordID, "system", action, message, r.clock.Now())
	if err := r.store.SaveAuditEvent(event); err != nil {
		return event, err
	}
	record, err := r.store.GetSyncRecord(recordID)
	if err != nil {
		return event, err
	}
	record.AddAudit(event.ID)
	return event, r.store.SaveSyncRecord(record)
}
