package domain

import (
	"sort"
	"strings"
)

type DevicePatch struct {
	Status     *DeviceStatus
	Label      *string
	Attributes map[string]string
}

func ApplyPatch(device Device, patch DevicePatch) Device {
	updated := device.Clone()
	if patch.Status != nil {
		updated.Status = *patch.Status
	}
	if patch.Label != nil {
		updated.Label = strings.TrimSpace(*patch.Label)
	}
	if updated.Attributes == nil {
		updated.Attributes = make(map[string]string)
	}
	for key, value := range patch.Attributes {
		if strings.TrimSpace(key) != "" {
			updated.Attributes[key] = value
		}
	}
	updated.Version++
	return updated
}

func SortDevices(devices []Device) []Device {
	result := append([]Device(nil), devices...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Vendor == result[j].Vendor {
			return result[i].ID < result[j].ID
		}
		return result[i].Vendor < result[j].Vendor
	})
	return result
}

func FindDevice(devices []Device, id string) (Device, bool) {
	for _, device := range devices {
		if device.ID == id {
			return device, true
		}
	}
	return Device{}, false
}

func CountStatuses(devices []Device) map[DeviceStatus]int {
	counts := make(map[DeviceStatus]int)
	for _, device := range devices {
		counts[device.Status]++
	}
	return counts
}

func CountKinds(devices []Device) map[DeviceKind]int {
	counts := make(map[DeviceKind]int)
	for _, device := range devices {
		counts[device.Kind]++
	}
	return counts
}

func SameVendor(devices []Device, vendor string) []Device {
	result := make([]Device, 0)
	for _, device := range devices {
		if device.Vendor == vendor {
			result = append(result, device)
		}
	}
	return SortDevices(result)
}

func TransitionAllowed(from, to SyncState) bool {
	switch from {
	case SyncPending:
		return to == SyncRunning || to == SyncFailed
	case SyncRunning:
		return to == SyncSuccess || to == SyncFailed
	case SyncFailed:
		return to == SyncRunning || to == SyncFailed
	case SyncSuccess:
		return to == SyncSuccess
	default:
		return false
	}
}

func ApplyState(record SyncRecord, next SyncState) SyncRecord {
	if TransitionAllowed(record.State, next) {
		record.State = next
		record.Revision++
	}
	return record
}

func MergeAttributes(base, overlay map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(overlay))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range overlay {
		merged[key] = value
	}
	return merged
}
