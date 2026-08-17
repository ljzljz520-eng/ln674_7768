package service

import (
	"errors"
	"fmt"
	"time"

	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

var ErrConfirmationUnavailable = errors.New("sync record cannot be confirmed")

type ConfirmationService struct {
	store   *storage.Store
	clock   Clock
	barrier <-chan struct{}
	ready   chan<- struct{}
}

func NewConfirmationService(store *storage.Store, clock Clock, barrier <-chan struct{}, ready chan<- struct{}) *ConfirmationService {
	return &ConfirmationService{store: store, clock: clock, barrier: barrier, ready: ready}
}

func (c *ConfirmationService) Confirm(recordID, operator string) (domain.AuditEvent, error) {
	record, err := c.store.GetSyncRecord(recordID)
	if err != nil {
		return domain.AuditEvent{}, err
	}
	if !record.CanConfirm() || operator == "" {
		return domain.AuditEvent{}, ErrConfirmationUnavailable
	}
	if c.ready != nil {
		c.ready <- struct{}{}
	}
	if c.barrier != nil {
		<-c.barrier
	}
	record.Confirmations++
	eventID := fmt.Sprintf("confirm:%s:%s", recordID, operator)
	record.AddAudit(eventID)
	event := domain.NewAuditEvent(eventID, recordID, operator, domain.ActionConfirmed, "operator confirmed synchronized status", c.clock.Now())
	if err := c.store.SaveConfirmation(record, event); err != nil {
		return domain.AuditEvent{}, err
	}
	return event, nil
}

func (c *ConfirmationService) ConfirmedAt() time.Time {
	return c.clock.Now()
}
