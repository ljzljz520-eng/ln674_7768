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
	// Pre-read the record so the operator can be validated up front and so
	// concurrent callers can coordinate on the barrier before any mutation.
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
	eventID := fmt.Sprintf("confirm:%s:%s", recordID, operator)
	event := domain.NewAuditEvent(eventID, recordID, operator, domain.ActionConfirmed, "operator confirmed synchronized status", c.clock.Now())
	// ApplyConfirmation reads the current record, bumps the confirmation count,
	// appends this event, and writes both back in a single transaction. The
	// read and the write share one transaction, so two operators confirming at
	// once cannot overwrite each other's update.
	if _, err := c.store.ApplyConfirmation(recordID, event); err != nil {
		return domain.AuditEvent{}, err
	}
	return event, nil
}

func (c *ConfirmationService) ConfirmedAt() time.Time {
	return c.clock.Now()
}
