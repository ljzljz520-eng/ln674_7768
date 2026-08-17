package storage

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"campus-device-hub/internal/domain"
	"go.etcd.io/bbolt"
)

type IntegrityReport struct {
	Devices       int
	Records       int
	Events        int
	Retries       int
	BrokenLinks   int
	InvalidStates int
	Warnings      []string
}

func (s *Store) InspectIntegrity() (IntegrityReport, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return IntegrityReport{}, err
	}
	records, err := s.ListSyncRecords()
	if err != nil {
		return IntegrityReport{}, err
	}
	events, err := s.ListAuditEvents()
	if err != nil {
		return IntegrityReport{}, err
	}
	tasks, err := s.ListRetryTasks()
	if err != nil {
		return IntegrityReport{}, err
	}
	report := IntegrityReport{Devices: len(devices), Records: len(records), Events: len(events), Retries: len(tasks), Warnings: []string{}}
	recordMap := make(map[string]domain.SyncRecord, len(records))
	for _, record := range records {
		recordMap[record.ID] = record
		if len(domain.ValidateRecord(record)) > 0 {
			report.InvalidStates++
		}
	}
	for _, device := range devices {
		if len(domain.ValidateDevice(device)) > 0 {
			report.InvalidStates++
		}
	}
	for _, event := range events {
		if _, ok := recordMap[event.RecordID]; !ok {
			report.BrokenLinks++
			report.Warnings = append(report.Warnings, "audit event references missing record: "+event.ID)
		}
	}
	for _, task := range tasks {
		if _, ok := recordMap[task.RecordID]; !ok {
			report.BrokenLinks++
			report.Warnings = append(report.Warnings, "retry task references missing record: "+task.ID)
		}
	}
	sort.Strings(report.Warnings)
	return report, nil
}

func (r IntegrityReport) Healthy() bool {
	return r.BrokenLinks == 0 && r.InvalidStates == 0
}

func (r IntegrityReport) WarningText() string {
	return strings.Join(r.Warnings, " | ")
}

func (s *Store) CountBucket(bucket []byte) (int, error) {
	count := 0
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return errors.New("bucket missing")
		}
		count = b.Stats().KeyN
		return nil
	})
	return count, err
}

func (s *Store) BucketCounts() (map[string]int, error) {
	counts := make(map[string]int, 4)
	for name, bucket := range map[string][]byte{"devices": devicesBucket, "records": recordsBucket, "events": auditBucket, "retries": retriesBucket} {
		count, err := s.CountBucket(bucket)
		if err != nil {
			return nil, err
		}
		counts[name] = count
	}
	return counts, nil
}

func (s *Store) LastActivity() (time.Time, error) {
	events, err := s.ListAuditEvents()
	if err != nil {
		return time.Time{}, err
	}
	if len(events) == 0 {
		return time.Time{}, nil
	}
	latest := events[0].CreatedAt
	for _, event := range events[1:] {
		if event.CreatedAt.After(latest) {
			latest = event.CreatedAt
		}
	}
	return latest, nil
}

func (s *Store) RecordTimeline(recordID string) ([]domain.AuditEvent, error) {
	events, err := s.ListAuditEventsForRecord(recordID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].ID < events[j].ID
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events, nil
}

func (s *Store) VerifyRecordLinks(recordID string) error {
	record, err := s.GetSyncRecord(recordID)
	if err != nil {
		return err
	}
	for _, eventID := range record.AuditEventIDs {
		if _, err := s.GetAuditEvent(eventID); err != nil {
			return fmt.Errorf("record %s references %s: %w", recordID, eventID, err)
		}
	}
	return nil
}
