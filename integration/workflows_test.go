package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"campus-device-hub/internal/adapters"
	"campus-device-hub/internal/dashboard"
	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/service"
	"campus-device-hub/internal/storage"
)

func TestWorkflowSyncDashboard(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	clock := service.NewFixedClock(time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC))
	syncer := service.NewSyncService(store, clock, adapters.DefaultFixtures())
	if _, err := syncer.SyncAll(); err != nil {
		t.Fatal(err)
	}
	summary, err := dashboard.NewView(store).Summary()
	if err != nil || summary.TotalDevices != 5 || summary.TotalSyncs != 3 {
		t.Fatalf("dashboard summary: %+v %v", summary, err)
	}
}

func TestWorkflowRetryAudit(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	clock := service.NewFixedClock(time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC))
	syncer := service.NewSyncService(store, clock, []adapters.VendorAdapter{adapters.BroadcastVendor{Fixture: adapters.Fixture{VendorName: "CampusCast", DeviceKind: domain.KindBroadcast, Failure: "timeout"}}})
	record, err := syncer.SyncVendor(context.Background(), "CampusCast")
	if err != nil || record.State != domain.SyncFailed {
		t.Fatalf("failed sync: %+v %v", record, err)
	}
	clock.Advance(2 * time.Minute)
	if err := syncer.RetryService().Run(context.Background(), func(context.Context, domain.RetryTask) error { return nil }); err != nil {
		t.Fatal(err)
	}
	events, err := syncer.AuditService().ForRecord(record.ID)
	if err != nil || len(events) != 3 {
		t.Fatalf("audit chain: %+v %v", events, err)
	}
}

func TestWorkflowConcurrentConfirmation(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	record := domain.SyncRecord{ID: "sync:workflow", Vendor: "BrightGrid", State: domain.SyncSuccess, AuditEventIDs: []string{}}
	if err := store.SaveSyncRecord(record); err != nil {
		t.Fatal(err)
	}
	barrier := make(chan struct{})
	ready := make(chan struct{}, 2)
	confirm := service.NewConfirmationService(store, service.NewFixedClock(time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)), barrier, ready)
	var group sync.WaitGroup
	for _, operator := range []string{"operator-a", "operator-b"} {
		group.Add(1)
		go func(name string) {
			defer group.Done()
			if _, err := confirm.Confirm(record.ID, name); err != nil {
				t.Errorf("confirm %s: %v", name, err)
			}
		}(operator)
	}
	<-ready
	<-ready
	close(barrier)
	group.Wait()
	loaded, err := store.GetSyncRecord(record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Confirmations != 2 || len(loaded.AuditEventIDs) != 2 {
		t.Fatalf("confirmation workflow lost state: %+v", loaded)
	}
}
