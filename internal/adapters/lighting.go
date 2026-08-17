package adapters

import (
	"context"
	"time"

	"campus-device-hub/internal/domain"
)

type LightingVendor struct{ Fixture }

func (l LightingVendor) Fetch(ctx context.Context, now time.Time) ([]domain.Device, error) {
	devices, err := l.Fixture.Fetch(ctx, now)
	if err != nil {
		return nil, err
	}
	for index := range devices {
		devices[index].Attributes["protocol"] = "dmx-json"
	}
	return devices, nil
}

func NormalizeLighting(device domain.Device) domain.Device {
	device.Kind = domain.KindLighting
	if device.Status == "ready" {
		device.Status = domain.StatusOnline
	}
	if device.Status == "fault" {
		device.Status = domain.StatusOffline
	}
	return device
}
