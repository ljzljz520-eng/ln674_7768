package storage

import (
	"testing"
	"time"

	"campus-device-hub/internal/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := t.TempDir() + "/campus.db"
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	record := domain.SyncRecord{ID: "sync:reopen", Vendor: "NorthGate", State: domain.SyncSuccess, AuditEventIDs: []string{"event:reopen"}}
	device := domain.Device{ID: "device:reopen", Kind: domain.KindAccess, Vendor: "NorthGate", Status: domain.StatusOnline, UpdatedAt: time.Unix(20, 0)}
	event := domain.NewAuditEvent("event:reopen", record.ID, "operator", domain.ActionConfirmed, "confirmed", time.Unix(20, 0))
	if err := store.SaveDevice(device); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveSyncRecord(record); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAuditEvent(event); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	loaded, err := store.GetSyncRecord(record.ID)
	if err != nil || loaded.DeviceCount != record.DeviceCount || len(loaded.AuditEventIDs) != 1 {
		t.Fatalf("record did not survive reopen: %+v %v", loaded, err)
	}
	if _, err := store.GetDevice(device.ID); err != nil {
		t.Fatal(err)
	}
}

func TestStoreListsEntitiesDeterministically(t *testing.T) {
	store, err := Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for _, id := range []string{"b", "a"} {
		if err := store.SaveDevice(domain.Device{ID: id, Vendor: "v", Kind: domain.KindAccess}); err != nil {
			t.Fatal(err)
		}
	}
	devices, err := store.ListDevices()
	if err != nil || len(devices) != 2 || devices[0].ID != "a" {
		t.Fatalf("unexpected devices: %+v %v", devices, err)
	}
}
