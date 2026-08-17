package service

import (
	"fmt"
	"sort"
	"time"

	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

type AuditService struct {
	store *storage.Store
	clock Clock
}

func NewAuditService(store *storage.Store, clock Clock) *AuditService {
	return &AuditService{store: store, clock: clock}
}

func (a *AuditService) Record(recordID, operator string, action domain.AuditAction, message string) (domain.AuditEvent, error) {
	at := a.clock.Now()
	id := fmt.Sprintf("audit:%s:%s:%s:%d", recordID, operator, action, at.UnixNano())
	event := domain.NewAuditEvent(id, recordID, operator, action, message, at)
	if err := a.store.SaveAuditEvent(event); err != nil {
		return domain.AuditEvent{}, err
	}
	if recordID != "" {
		record, err := a.store.GetSyncRecord(recordID)
		if err != nil {
			return domain.AuditEvent{}, err
		}
		record.AddAudit(event.ID)
		if err := a.store.SaveSyncRecord(record); err != nil {
			return domain.AuditEvent{}, err
		}
	}
	return event, nil
}

func (a *AuditService) ForRecord(recordID string) ([]domain.AuditEvent, error) {
	return a.store.ListAuditEventsForRecord(recordID)
}

func (a *AuditService) All() ([]domain.AuditEvent, error) {
	events, err := a.store.ListAuditEvents()
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool { return events[i].CreatedAt.Before(events[j].CreatedAt) })
	return events, nil
}

func FormatEventTime(event domain.AuditEvent) string {
	return event.CreatedAt.UTC().Format(time.RFC3339)
}
