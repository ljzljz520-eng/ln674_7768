package adapters

import (
	"context"
	"time"

	"campus-device-hub/internal/domain"
)

type VendorAdapter interface {
	Name() string
	Kind() domain.DeviceKind
	Fetch(context.Context, time.Time) ([]domain.Device, error)
}

type Fixture struct {
	VendorName string
	DeviceKind domain.DeviceKind
	Devices    []domain.Device
	Failure    string
}

func (f Fixture) Name() string { return f.VendorName }

func (f Fixture) Kind() domain.DeviceKind { return f.DeviceKind }

func (f Fixture) Fetch(ctx context.Context, now time.Time) ([]domain.Device, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if f.Failure != "" {
		return nil, VendorError{Vendor: f.VendorName, Message: f.Failure}
	}
	devices := make([]domain.Device, 0, len(f.Devices))
	for _, device := range f.Devices {
		copyDevice := device.Clone()
		copyDevice.UpdatedAt = now
		devices = append(devices, copyDevice)
	}
	return devices, nil
}

type VendorError struct {
	Vendor  string
	Message string
}

func (e VendorError) Error() string {
	return e.Vendor + ": " + e.Message
}

func DefaultFixtures() []VendorAdapter {
	return []VendorAdapter{
		AccessVendor{Fixture: Fixture{VendorName: "NorthGate", DeviceKind: domain.KindAccess, Devices: []domain.Device{
			{ID: "access-101", Kind: domain.KindAccess, Vendor: "NorthGate", Label: "Library Entry", Status: domain.StatusOnline, Attributes: map[string]string{"zone": "library"}},
			{ID: "access-102", Kind: domain.KindAccess, Vendor: "NorthGate", Label: "Science Entry", Status: domain.StatusWarning, Attributes: map[string]string{"zone": "science"}},
		}}},
		BroadcastVendor{Fixture: Fixture{VendorName: "CampusCast", DeviceKind: domain.KindBroadcast, Devices: []domain.Device{
			{ID: "broadcast-201", Kind: domain.KindBroadcast, Vendor: "CampusCast", Label: "Main Quad", Status: domain.StatusOnline, Attributes: map[string]string{"channel": "emergency"}},
			{ID: "broadcast-202", Kind: domain.KindBroadcast, Vendor: "CampusCast", Label: "Dormitory", Status: domain.StatusOnline, Attributes: map[string]string{"channel": "general"}},
		}}},
		LightingVendor{Fixture: Fixture{VendorName: "BrightGrid", DeviceKind: domain.KindLighting, Devices: []domain.Device{
			{ID: "light-301", Kind: domain.KindLighting, Vendor: "BrightGrid", Label: "Walkway A", Status: domain.StatusOffline, Attributes: map[string]string{"circuit": "A"}},
		}}},
	}
}
