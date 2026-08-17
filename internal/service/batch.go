package service

import (
	"context"
	"fmt"
	"sort"

	"campus-device-hub/internal/domain"
)

type BatchReport struct {
	Records  []domain.SyncRecord
	Success  int
	Failures int
	Devices  int
}

func (s *SyncService) SyncSelected(ctx context.Context, vendors []string) (BatchReport, error) {
	names := append([]string(nil), vendors...)
	sort.Strings(names)
	report := BatchReport{Records: make([]domain.SyncRecord, 0, len(names))}
	for _, vendor := range names {
		if vendor == "" {
			continue
		}
		record, err := s.SyncVendor(ctx, vendor)
		if err != nil {
			return report, err
		}
		report.Records = append(report.Records, record)
		if record.State == domain.SyncSuccess {
			report.Success++
			report.Devices += record.DeviceCount
		} else {
			report.Failures++
		}
	}
	return report, nil
}

func (s *SyncService) ValidateStoredState() ([]domain.ValidationIssue, error) {
	devices, err := s.store.ListDevices()
	if err != nil {
		return nil, err
	}
	records, err := s.store.ListSyncRecords()
	if err != nil {
		return nil, err
	}
	events, err := s.store.ListAuditEvents()
	if err != nil {
		return nil, err
	}
	tasks, err := s.store.ListRetryTasks()
	if err != nil {
		return nil, err
	}
	issues := make([]domain.ValidationIssue, 0)
	for _, device := range devices {
		issues = append(issues, domain.ValidateDevice(device)...)
	}
	for _, record := range records {
		issues = append(issues, domain.ValidateRecord(record)...)
	}
	for _, event := range events {
		issues = append(issues, domain.ValidateAuditEvent(event)...)
	}
	for _, task := range tasks {
		issues = append(issues, domain.ValidateRetryTask(task)...)
	}
	return issues, nil
}

func (s *SyncService) Reconcile(ctx context.Context, vendor string) (domain.SyncRecord, error) {
	if vendor == "" {
		return domain.SyncRecord{}, fmt.Errorf("vendor is required")
	}
	records, err := s.store.RecordsForVendor(vendor)
	if err != nil {
		return domain.SyncRecord{}, err
	}
	if len(records) > 0 {
		latest := records[len(records)-1]
		if latest.State == domain.SyncSuccess {
			return latest, nil
		}
	}
	return s.SyncVendor(ctx, vendor)
}

func (s *SyncService) LatestRecord(vendor string) (domain.SyncRecord, error) {
	records, err := s.store.RecordsForVendor(vendor)
	if err != nil {
		return domain.SyncRecord{}, err
	}
	if len(records) == 0 {
		return domain.SyncRecord{}, fmt.Errorf("no record for %s", vendor)
	}
	return records[len(records)-1], nil
}
