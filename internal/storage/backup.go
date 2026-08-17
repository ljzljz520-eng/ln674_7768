package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"campus-device-hub/internal/domain"
)

type Backup struct {
	CreatedAt time.Time
	Devices   []domain.Device
	Records   []domain.SyncRecord
	Events    []domain.AuditEvent
	Retries   []domain.RetryTask
}

func (s *Store) CreateBackup(at time.Time) (Backup, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return Backup{}, err
	}
	records, err := s.ListSyncRecords()
	if err != nil {
		return Backup{}, err
	}
	events, err := s.ListAuditEvents()
	if err != nil {
		return Backup{}, err
	}
	retries, err := s.ListRetryTasks()
	if err != nil {
		return Backup{}, err
	}
	return Backup{CreatedAt: at, Devices: domain.SortDevices(devices), Records: records, Events: events, Retries: retries}, nil
}

func (b Backup) Marshal() ([]byte, error) {
	return json.MarshalIndent(b, "", "  ")
}

func ParseBackup(data []byte) (Backup, error) {
	var backup Backup
	if len(bytes.TrimSpace(data)) == 0 {
		return backup, fmt.Errorf("backup is empty")
	}
	if err := json.Unmarshal(data, &backup); err != nil {
		return backup, fmt.Errorf("parse backup: %w", err)
	}
	return backup, nil
}

func (b Backup) Valid() bool {
	if b.CreatedAt.IsZero() {
		return false
	}
	for _, device := range b.Devices {
		if len(domain.ValidateDevice(device)) > 0 {
			return false
		}
	}
	for _, record := range b.Records {
		if len(domain.ValidateRecord(record)) > 0 {
			return false
		}
	}
	for _, event := range b.Events {
		if len(domain.ValidateAuditEvent(event)) > 0 {
			return false
		}
	}
	for _, task := range b.Retries {
		if len(domain.ValidateRetryTask(task)) > 0 {
			return false
		}
	}
	return true
}

func (b Backup) DeviceIDs() []string {
	ids := make([]string, 0, len(b.Devices))
	for _, device := range b.Devices {
		ids = append(ids, device.ID)
	}
	sort.Strings(ids)
	return ids
}

func (b Backup) RecordIDs() []string {
	ids := make([]string, 0, len(b.Records))
	for _, record := range b.Records {
		ids = append(ids, record.ID)
	}
	sort.Strings(ids)
	return ids
}

func (b Backup) EventCountFor(recordID string) int {
	count := 0
	for _, event := range b.Events {
		if event.RecordID == recordID {
			count++
		}
	}
	return count
}

func (b Backup) PendingRetryCount() int {
	count := 0
	for _, task := range b.Retries {
		if !task.Resolved {
			count++
		}
	}
	return count
}

func (b Backup) Summary() string {
	return fmt.Sprintf("backup at %s: devices=%d records=%d events=%d retries=%d", b.CreatedAt.UTC().Format(time.RFC3339), len(b.Devices), len(b.Records), len(b.Events), len(b.Retries))
}
