package service

import (
	"context"
	"testing"
	"time"

	"campus-device-hub/internal/adapters"
	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

func TestRetryRunResolvesDueTask(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	clock := NewFixedClock(time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC))
	service := NewSyncService(store, clock, []adapters.VendorAdapter{adapters.AccessVendor{Fixture: adapters.Fixture{VendorName: "Broken", DeviceKind: domain.KindAccess, Failure: "offline"}}})
	if _, err := service.SyncVendor(context.Background(), "Broken"); err != nil {
		t.Fatal(err)
	}
	clock.Advance(2 * time.Minute)
	if err := service.RetryService().Run(context.Background(), func(context.Context, domain.RetryTask) error { return nil }); err != nil {
		t.Fatal(err)
	}
	tasks, err := store.ListRetryTasks()
	if err != nil || len(tasks) != 1 || !tasks[0].Resolved {
		t.Fatalf("retry not resolved: %+v %v", tasks, err)
	}
}
