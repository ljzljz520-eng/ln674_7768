package storage

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"campus-device-hub/internal/domain"
)

type DeviceQuery struct {
	Vendor string
	Kind   domain.DeviceKind
	Status domain.DeviceStatus
	Since  time.Time
	Text   string
}

func (s *Store) QueryDevices(query DeviceQuery) ([]domain.Device, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return nil, err
	}
	result := make([]domain.Device, 0, len(devices))
	for _, device := range devices {
		if query.Vendor != "" && device.Vendor != query.Vendor {
			continue
		}
		if query.Kind != "" && device.Kind != query.Kind {
			continue
		}
		if query.Status != "" && device.Status != query.Status {
			continue
		}
		if !query.Since.IsZero() && device.UpdatedAt.Before(query.Since) {
			continue
		}
		if query.Text != "" && !strings.Contains(strings.ToLower(device.Label), strings.ToLower(query.Text)) {
			continue
		}
		result = append(result, device)
	}
	return domain.SortDevices(result), nil
}

func (s *Store) RecordsForVendor(vendor string) ([]domain.SyncRecord, error) {
	records, err := s.ListSyncRecords()
	if err != nil {
		return nil, err
	}
	result := make([]domain.SyncRecord, 0)
	for _, record := range records {
		if vendor == "" || record.Vendor == vendor {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartedAt.Before(result[j].StartedAt) })
	return result, nil
}

func (s *Store) RecordsWithState(state domain.SyncState) ([]domain.SyncRecord, error) {
	records, err := s.ListSyncRecords()
	if err != nil {
		return nil, err
	}
	result := make([]domain.SyncRecord, 0)
	for _, record := range records {
		if record.State == state {
			result = append(result, record)
		}
	}
	return result, nil
}

func (s *Store) AuditForOperator(operator string) ([]domain.AuditEvent, error) {
	events, err := s.ListAuditEvents()
	if err != nil {
		return nil, err
	}
	result := make([]domain.AuditEvent, 0)
	for _, event := range events {
		if operator == "" || event.Operator == operator {
			result = append(result, event)
		}
	}
	return result, nil
}

func (s *Store) AuditSince(since time.Time) ([]domain.AuditEvent, error) {
	events, err := s.ListAuditEvents()
	if err != nil {
		return nil, err
	}
	result := make([]domain.AuditEvent, 0)
	for _, event := range events {
		if since.IsZero() || !event.CreatedAt.Before(since) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (s *Store) PendingRetries() ([]domain.RetryTask, error) {
	tasks, err := s.ListRetryTasks()
	if err != nil {
		return nil, err
	}
	result := make([]domain.RetryTask, 0)
	for _, task := range tasks {
		if !task.Resolved {
			result = append(result, task)
		}
	}
	return result, nil
}

func (s *Store) MarshalSnapshot() ([]byte, error) {
	devices, records, err := s.Snapshot()
	if err != nil {
		return nil, err
	}
	events, err := s.ListAuditEvents()
	if err != nil {
		return nil, err
	}
	tasks, err := s.ListRetryTasks()
	if err != nil {
		return nil, err
	}
	payload := struct {
		Devices map[string]domain.Device
		Records map[string]domain.SyncRecord
		Events  []domain.AuditEvent
		Retries []domain.RetryTask
	}{devices, records, events, tasks}
	return json.MarshalIndent(payload, "", "  ")
}
