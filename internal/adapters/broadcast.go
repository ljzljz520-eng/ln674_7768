package adapters

import (
	"context"
	"time"

	"campus-device-hub/internal/domain"
)

type BroadcastVendor struct{ Fixture }

func (b BroadcastVendor) Fetch(ctx context.Context, now time.Time) ([]domain.Device, error) {
	devices, err := b.Fixture.Fetch(ctx, now)
	if err != nil {
		return nil, err
	}
	for index := range devices {
		devices[index].Attributes["protocol"] = "tone-xml"
	}
	return devices, nil
}

func NormalizeBroadcast(device domain.Device) domain.Device {
	device.Kind = domain.KindBroadcast
	if device.Status == "ready" {
		device.Status = domain.StatusOnline
	}
	if device.Status == "muted" {
		device.Status = domain.StatusWarning
	}
	return device
}
