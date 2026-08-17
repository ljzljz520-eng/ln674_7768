package adapters

import (
	"context"
	"time"

	"campus-device-hub/internal/domain"
)

type AccessVendor struct{ Fixture }

func (a AccessVendor) Fetch(ctx context.Context, now time.Time) ([]domain.Device, error) {
	devices, err := a.Fixture.Fetch(ctx, now)
	if err != nil {
		return nil, err
	}
	for index := range devices {
		devices[index].Attributes["protocol"] = "badge-v2"
	}
	return devices, nil
}

func NormalizeAccess(device domain.Device) domain.Device {
	device.Kind = domain.KindAccess
	if device.Status == "ready" {
		device.Status = domain.StatusOnline
	}
	if device.Status == "blocked" {
		device.Status = domain.StatusOffline
	}
	return device
}
