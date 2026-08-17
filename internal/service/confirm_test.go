package service

import (
	"sync"
	"testing"
	"time"

	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

func TestConcurrentConfirmationKeepsBothOperators(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	record := domain.SyncRecord{ID: "sync:concurrent", Vendor: "NorthGate", State: domain.SyncSuccess, AuditEventIDs: []string{}}
	if err := store.SaveSyncRecord(record); err != nil {
		t.Fatal(err)
	}
	barrier := make(chan struct{})
	ready := make(chan struct{}, 2)
	clock := NewFixedClock(time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC))
	service := NewConfirmationService(store, clock, barrier, ready)
	var group sync.WaitGroup
	errors := make(chan error, 2)
	for _, operator := range []string{"alice", "bob"} {
		group.Add(1)
		go func(name string) {
			defer group.Done()
			_, confirmErr := service.Confirm(record.ID, name)
			errors <- confirmErr
		}(operator)
	}
	<-ready
	<-ready
	close(barrier)
	group.Wait()
	close(errors)
	for confirmErr := range errors {
		if confirmErr != nil {
			t.Fatal(confirmErr)
		}
	}
	loaded, err := store.GetSyncRecord(record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Confirmations != 2 {
		t.Fatalf("confirmation count = %d", loaded.Confirmations)
	}
	events, err := store.ListAuditEventsForRecord(record.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("audit events = %d", len(events))
	}
}
