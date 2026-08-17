package storage

import (
	"encoding/json"
	"sort"

	"campus-device-hub/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveConfirmation(record domain.SyncRecord, event domain.AuditEvent) error {
	recordID := record.ID
	eventData, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		records := tx.Bucket(recordsBucket)
		recordData := records.Get([]byte(recordID))
		if recordData == nil {
			return ErrNotFound
		}
		var record domain.SyncRecord
		if err := json.Unmarshal(recordData, &record); err != nil {
			return err
		}
		record.Confirmations++
		record.AddAudit(event.ID)
		recordData, err = json.Marshal(record)
		if err != nil {
			return err
		}
		if err := tx.Bucket(auditBucket).Put([]byte(event.ID), eventData); err != nil {
			return err
		}
		return records.Put([]byte(recordID), recordData)
	})
}

func (s *Store) ListAuditEventsForRecord(recordID string) ([]domain.AuditEvent, error) {
	record, err := s.GetSyncRecord(recordID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.AuditEvent, 0, len(record.AuditEventIDs))
	for _, id := range record.AuditEventIDs {
		event, getErr := s.GetAuditEvent(id)
		if getErr != nil {
			return nil, getErr
		}
		result = append(result, event)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (s *Store) Snapshot() (map[string]domain.Device, map[string]domain.SyncRecord, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return nil, nil, err
	}
	records, err := s.ListSyncRecords()
	if err != nil {
		return nil, nil, err
	}
	deviceMap := make(map[string]domain.Device, len(devices))
	recordMap := make(map[string]domain.SyncRecord, len(records))
	for _, device := range devices {
		deviceMap[device.ID] = device
	}
	for _, record := range records {
		recordMap[record.ID] = record
	}
	return deviceMap, recordMap, nil
}
