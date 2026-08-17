package dashboard

import (
	"testing"
	"time"

	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

func TestViewRendersSummaryAndCards(t *testing.T) {
	store, err := storage.Open(t.TempDir() + "/campus.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for _, device := range []domain.Device{{ID: "a", Vendor: "NorthGate", Kind: domain.KindAccess, Status: domain.StatusOnline}, {ID: "b", Vendor: "NorthGate", Kind: domain.KindAccess, Status: domain.StatusOffline}} {
		if err := store.SaveDevice(device); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.SaveSyncRecord(domain.SyncRecord{ID: "r", Vendor: "NorthGate", State: domain.SyncSuccess, FinishedAt: time.Unix(1, 0)}); err != nil {
		t.Fatal(err)
	}
	view := NewView(store)
	summary, err := view.Summary()
	if err != nil || summary.TotalDevices != 2 || summary.HealthyDevices != 1 || summary.HealthPercent() != 50 {
		t.Fatalf("summary: %+v %v", summary, err)
	}
	cards, err := view.VendorCards()
	if err != nil || len(cards) != 1 || cards[0].DeviceCount != 2 {
		t.Fatalf("cards: %+v %v", cards, err)
	}
}
