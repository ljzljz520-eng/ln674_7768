package storage

import (
	"encoding/json"
	"fmt"
	"sort"

	"campus-device-hub/internal/domain"
	"go.etcd.io/bbolt"
)

// ApplyConfirmation atomically records an operator's confirmation against the
// current stored record. It reads the record, increments its confirmation
// count, appends the audit-event reference, bumps the revision, writes the
// audit event, and writes the record back — all within a single bbolt
// transaction. Because the read and the write share one transaction, two
// operators confirming the same record at once are serialized at the storage
// layer and neither can overwrite the other's update (a lost-update that
// previously dropped one confirmation and its audit event).
func (s *Store) ApplyConfirmation(recordID string, event domain.AuditEvent) (domain.SyncRecord, error) {
	var updated domain.SyncRecord
	if recordID == "" {
		return updated, ErrNotFound
	}
	eventData, err := json.Marshal(event)
	if err != nil {
		return updated, fmt.Errorf("encode event %s: %w", event.ID, err)
	}
	err = s.db.Update(func(tx *bbolt.Tx) error {
		records := tx.Bucket(recordsBucket)
		if records == nil {
			return ErrNotFound
		}
		current := records.Get([]byte(recordID))
		if current == nil {
			return ErrNotFound
		}
		if err := decodeValue(current, &updated); err != nil {
			return err
		}
		updated.Confirmations++
		updated.AddAudit(event.ID)
		updated.Revision++
		recordData, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("encode record %s: %w", recordID, err)
		}
		if err := tx.Bucket(auditBucket).Put([]byte(event.ID), eventData); err != nil {
			return err
		}
		return records.Put([]byte(recordID), recordData)
	})
	return updated, err
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
