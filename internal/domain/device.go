package domain

import "time"

type DeviceKind string

const (
	KindAccess    DeviceKind = "access"
	KindBroadcast DeviceKind = "broadcast"
	KindLighting  DeviceKind = "lighting"
)

type DeviceStatus string

const (
	StatusOnline  DeviceStatus = "online"
	StatusOffline DeviceStatus = "offline"
	StatusWarning DeviceStatus = "warning"
	StatusUnknown DeviceStatus = "unknown"
)

type Device struct {
	ID         string            `json:"id"`
	Kind       DeviceKind        `json:"kind"`
	Vendor     string            `json:"vendor"`
	Status     DeviceStatus      `json:"status"`
	Label      string            `json:"label"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Version    int               `json:"version"`
}

func (d Device) IsHealthy() bool {
	return d.Status == StatusOnline
}

func (d Device) IsKnown() bool {
	return d.ID != "" && d.Kind != "" && d.Vendor != ""
}

func (d Device) DisplayStatus() string {
	switch d.Status {
	case StatusOnline:
		return "Online"
	case StatusOffline:
		return "Offline"
	case StatusWarning:
		return "Warning"
	default:
		return "Unknown"
	}
}

func (d Device) Clone() Device {
	copyDevice := d
	copyDevice.Attributes = make(map[string]string, len(d.Attributes))
	for key, value := range d.Attributes {
		copyDevice.Attributes[key] = value
	}
	return copyDevice
}
