package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"campus-device-hub/internal/domain"
	"go.etcd.io/bbolt"
)

var (
	devicesBucket = []byte("devices")
	recordsBucket = []byte("sync_records")
	auditBucket   = []byte("audit_events")
	retriesBucket = []byte("retry_tasks")
)

var ErrNotFound = errors.New("record not found")

type Store struct {
	db *bbolt.DB
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage path is required")
	}
	if err := ensureParent(path); err != nil {
		return nil, err
	}
	db, err := bbolt.Open(path, 0o600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bbolt: %w", err)
	}
	store := &Store{db: db}
	if err := db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{devicesBucket, recordsBucket, auditBucket, retriesBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("create buckets: %w", err)
	}
	return store, nil
}

func ensureParent(path string) error {
	parent := filepath.Dir(path)
	if parent == "." {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) SaveDevice(device domain.Device) error {
	return s.put(devicesBucket, device.ID, device)
}

func (s *Store) GetDevice(id string) (domain.Device, error) {
	var device domain.Device
	err := s.get(devicesBucket, id, &device)
	return device, err
}

func (s *Store) ListDevices() ([]domain.Device, error) {
	values, err := s.list(devicesBucket)
	if err != nil {
		return nil, err
	}
	devices := make([]domain.Device, 0, len(values))
	for _, value := range values {
		var device domain.Device
		if err := decodeValue(value, &device); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].ID < devices[j].ID })
	return devices, err
}

func (s *Store) SaveSyncRecord(record domain.SyncRecord) error {
	return s.put(recordsBucket, record.ID, record)
}

func (s *Store) GetSyncRecord(id string) (domain.SyncRecord, error) {
	var record domain.SyncRecord
	err := s.get(recordsBucket, id, &record)
	return record, err
}

func (s *Store) ListSyncRecords() ([]domain.SyncRecord, error) {
	values, err := s.list(recordsBucket)
	if err != nil {
		return nil, err
	}
	records := make([]domain.SyncRecord, 0, len(values))
	for _, value := range values {
		var record domain.SyncRecord
		if err := decodeValue(value, &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	return records, err
}

func (s *Store) SaveAuditEvent(event domain.AuditEvent) error {
	return s.put(auditBucket, event.ID, event)
}

func (s *Store) GetAuditEvent(id string) (domain.AuditEvent, error) {
	var event domain.AuditEvent
	err := s.get(auditBucket, id, &event)
	return event, err
}

func (s *Store) ListAuditEvents() ([]domain.AuditEvent, error) {
	values, err := s.list(auditBucket)
	if err != nil {
		return nil, err
	}
	events := make([]domain.AuditEvent, 0, len(values))
	for _, value := range values {
		var event domain.AuditEvent
		if err := decodeValue(value, &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].ID < events[j].ID
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events, err
}

func (s *Store) SaveRetryTask(task domain.RetryTask) error {
	return s.put(retriesBucket, task.ID, task)
}

func (s *Store) ListRetryTasks() ([]domain.RetryTask, error) {
	values, err := s.list(retriesBucket)
	if err != nil {
		return nil, err
	}
	tasks := make([]domain.RetryTask, 0, len(values))
	for _, value := range values {
		var task domain.RetryTask
		if err := decodeValue(value, &task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks, err
}

func (s *Store) put(bucket []byte, key string, value any) error {
	if key == "" {
		return errors.New("storage key is required")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", key, err)
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Put([]byte(key), data)
	})
}

func (s *Store) get(bucket []byte, key string, target any) error {
	if key == "" {
		return ErrNotFound
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucket).Get([]byte(key))
		if value == nil {
			return ErrNotFound
		}
		return json.Unmarshal(value, target)
	})
}

func (s *Store) list(bucket []byte) ([][]byte, error) {
	var values [][]byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).ForEach(func(_, value []byte) error {
			if value != nil {
				values = append(values, cloneBytes(value))
			}
			return nil
		})
	})
	return values, err
}
