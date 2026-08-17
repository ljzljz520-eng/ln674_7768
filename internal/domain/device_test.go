package domain

import "testing"

func TestDeviceHealthAndClone(t *testing.T) {
	device := Device{ID: "d1", Kind: KindAccess, Vendor: "v", Status: StatusOnline, Attributes: map[string]string{"zone": "a"}}
	if !device.IsKnown() || !device.IsHealthy() || device.DisplayStatus() != "Online" {
		t.Fatalf("unexpected device state: %+v", device)
	}
	clone := device.Clone()
	clone.Attributes["zone"] = "b"
	if device.Attributes["zone"] != "a" {
		t.Fatal("clone shares attributes")
	}
}

func TestDeviceDisplayStatuses(t *testing.T) {
	values := map[DeviceStatus]string{StatusOffline: "Offline", StatusWarning: "Warning", StatusUnknown: "Unknown"}
	for status, expected := range values {
		if (Device{Status: status}).DisplayStatus() != expected {
			t.Fatalf("status %s", status)
		}
	}
}
