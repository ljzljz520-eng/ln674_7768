package service

import (
	"context"
	"testing"
	"time"

	"campus-device-hub/internal/adapters"
	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

func TestSyncServicePersistsNormalizedDevices(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	clock := NewFixedClock(time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC))
	service := NewSyncService(store, clock, adapters.DefaultFixtures())
	record, err := service.SyncVendor(context.Background(), "NorthGate")
	if err != nil || record.State != domain.SyncSuccess || record.DeviceCount != 2 {
		t.Fatalf("sync failed: %+v %v", record, err)
	}
	devices, err := store.ListDevices()
	if err != nil || len(devices) != 2 {
		t.Fatalf("devices missing: %+v %v", devices, err)
	}
	events, err := service.AuditService().ForRecord(record.ID)
	if err != nil || len(events) != 2 {
		t.Fatalf("audit trail missing: %+v %v", events, err)
	}
}

func TestSyncServiceQueuesFailure(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	clock := NewFixedClock(time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC))
	service := NewSyncService(store, clock, []adapters.VendorAdapter{adapters.AccessVendor{Fixture: adapters.Fixture{VendorName: "Broken", DeviceKind: domain.KindAccess, Failure: "offline"}}})
	record, err := service.SyncVendor(context.Background(), "Broken")
	if err != nil || record.State != domain.SyncFailed {
		t.Fatalf("failure not recorded: %+v %v", record, err)
	}
	tasks, err := store.ListRetryTasks()
	if err != nil || len(tasks) != 1 || tasks[0].Vendor != "Broken" {
		t.Fatalf("retry missing: %+v %v", tasks, err)
	}
}
