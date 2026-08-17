package service

import (
	"context"
	"fmt"
	"sort"

	"campus-device-hub/internal/adapters"
	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

type SyncService struct {
	store   *storage.Store
	clock   Clock
	catalog *adapters.Catalog
	audit   *AuditService
	retry   *RetryService
}

func NewSyncService(store *storage.Store, clock Clock, vendors []adapters.VendorAdapter) *SyncService {
	audit := NewAuditService(store, clock)
	retry := NewRetryService(store, clock)
	return &SyncService{store: store, clock: clock, catalog: adapters.NewCatalog(vendors), audit: audit, retry: retry}
}

func (s *SyncService) SyncAll() ([]domain.SyncRecord, error) {
	names := s.catalog.Names()
	sort.Strings(names)
	result := make([]domain.SyncRecord, 0, len(names))
	for _, name := range names {
		record, err := s.SyncVendor(context.Background(), name)
		if err != nil {
			return result, err
		}
		result = append(result, record)
	}
	return result, nil
}

func (s *SyncService) SyncVendor(ctx context.Context, vendor string) (domain.SyncRecord, error) {
	at := s.clock.Now()
	record := domain.SyncRecord{ID: fmt.Sprintf("sync:%s:%d", vendor, at.UnixNano()), Vendor: vendor, State: domain.SyncPending, AuditEventIDs: []string{}}
	record.MarkStarted(at)
	if err := s.store.SaveSyncRecord(record); err != nil {
		return record, err
	}
	startedEvent, err := s.audit.Record(record.ID, "system", domain.ActionSyncStarted, "vendor synchronization started")
	if err != nil {
		return record, err
	}
	record.AddAudit(startedEvent.ID)
	devices, err := s.catalog.Fetch(ctx, vendor, at)
	if err != nil {
		record.MarkFailure(s.clock.Now(), err.Error())
		if saveErr := s.store.SaveSyncRecord(record); saveErr != nil {
			return record, saveErr
		}
		failedEvent, auditErr := s.audit.Record(record.ID, "system", domain.ActionSyncFailed, err.Error())
		if auditErr != nil {
			return record, auditErr
		}
		record.AddAudit(failedEvent.ID)
		if saveErr := s.store.SaveSyncRecord(record); saveErr != nil {
			return record, saveErr
		}
		if queueErr := s.retry.Queue(record, err.Error()); queueErr != nil {
			return record, queueErr
		}
		return record, nil
	}
	for _, device := range devices {
		device = normalizeByKind(device)
		if err := s.store.SaveDevice(device); err != nil {
			return record, err
		}
	}
	record.MarkSuccess(s.clock.Now(), len(devices))
	if err := s.store.SaveSyncRecord(record); err != nil {
		return record, err
	}
	completedEvent, err := s.audit.Record(record.ID, "system", domain.ActionSyncCompleted, fmt.Sprintf("%d devices synchronized", len(devices)))
	if err != nil {
		return record, err
	}
	record.AddAudit(completedEvent.ID)
	if err := s.store.SaveSyncRecord(record); err != nil {
		return record, err
	}
	return record, nil
}

func normalizeByKind(device domain.Device) domain.Device {
	switch device.Kind {
	case domain.KindAccess:
		return adapters.NormalizeAccess(device)
	case domain.KindBroadcast:
		return adapters.NormalizeBroadcast(device)
	case domain.KindLighting:
		return adapters.NormalizeLighting(device)
	default:
		device.Status = domain.StatusUnknown
		return device
	}
}

func (s *SyncService) RetryService() *RetryService {
	return s.retry
}

func (s *SyncService) AuditService() *AuditService {
	return s.audit
}
