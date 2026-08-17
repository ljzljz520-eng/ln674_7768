package adapters

import (
	"context"
	"testing"
	"time"

	"campus-device-hub/internal/domain"
)

func TestVendorAdaptersNormalizeDifferentProtocols(t *testing.T) {
	now := time.Date(2026, 2, 2, 9, 0, 0, 0, time.UTC)
	for _, adapter := range DefaultFixtures() {
		devices, err := adapter.Fetch(context.Background(), now)
		if err != nil || len(devices) == 0 {
			t.Fatalf("fetch %s: %v", adapter.Name(), err)
		}
		for _, device := range devices {
			if device.UpdatedAt != now || device.Attributes["protocol"] == "" {
				t.Fatalf("normalization failed: %+v", device)
			}
		}
	}
	if NormalizeAccess(domain.Device{Status: "blocked"}).Status != domain.StatusOffline {
		t.Fatal("access mapping failed")
	}
	if NormalizeBroadcast(domain.Device{Status: "muted"}).Status != domain.StatusWarning {
		t.Fatal("broadcast mapping failed")
	}
	if NormalizeLighting(domain.Device{Status: "fault"}).Status != domain.StatusOffline {
		t.Fatal("lighting mapping failed")
	}
}

func TestCatalogReportsUnknownVendor(t *testing.T) {
	catalog := NewCatalog(DefaultFixtures())
	if _, err := catalog.Fetch(context.Background(), "missing", time.Now()); err == nil {
		t.Fatal("expected missing vendor error")
	}
}
