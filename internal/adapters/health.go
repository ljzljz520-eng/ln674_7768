package adapters

import (
	"context"
	"sort"
	"strings"
	"time"

	"campus-device-hub/internal/domain"
)

type VendorHealth struct {
	Vendor       string
	Kind         domain.DeviceKind
	Reachable    bool
	DeviceCount  int
	HealthyCount int
	Message      string
}

func CheckHealth(ctx context.Context, adapter VendorAdapter, now time.Time) VendorHealth {
	health := VendorHealth{Vendor: adapter.Name(), Kind: adapter.Kind()}
	devices, err := adapter.Fetch(ctx, now)
	if err != nil {
		health.Message = err.Error()
		return health
	}
	health.Reachable = true
	health.DeviceCount = len(devices)
	for _, device := range devices {
		if device.IsHealthy() {
			health.HealthyCount++
		}
	}
	if health.DeviceCount == 0 {
		health.Message = "vendor returned no devices"
	} else if health.HealthyCount == health.DeviceCount {
		health.Message = "all devices healthy"
	} else {
		health.Message = "vendor reachable with device warnings"
	}
	return health
}

func CheckAll(ctx context.Context, adapters []VendorAdapter, now time.Time) []VendorHealth {
	result := make([]VendorHealth, 0, len(adapters))
	for _, adapter := range adapters {
		if adapter != nil {
			result = append(result, CheckHealth(ctx, adapter, now))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Vendor < result[j].Vendor })
	return result
}

func HealthLabel(health VendorHealth) string {
	if !health.Reachable {
		return "unreachable"
	}
	if health.HealthyCount == health.DeviceCount && health.DeviceCount > 0 {
		return "healthy"
	}
	return "degraded"
}

func ValidateVendorDevices(adapter VendorAdapter, devices []domain.Device) []domain.ValidationIssue {
	issues := make([]domain.ValidationIssue, 0)
	for index, device := range devices {
		if device.Vendor != adapter.Name() {
			issues = append(issues, domain.ValidationIssue{Field: "vendor", Message: adapter.Name() + " device mismatch"})
		}
		if device.Kind != adapter.Kind() {
			issues = append(issues, domain.ValidationIssue{Field: "kind", Message: adapter.Name() + " device mismatch"})
		}
		for _, issue := range domain.ValidateDevice(device) {
			issues = append(issues, domain.ValidationIssue{Field: strings.Join([]string{"devices", string(rune(index + '0')), issue.Field}, "."), Message: issue.Message})
		}
	}
	return issues
}

func DeviceIDs(devices []domain.Device) []string {
	ids := make([]string, 0, len(devices))
	for _, device := range devices {
		ids = append(ids, device.ID)
	}
	sort.Strings(ids)
	return ids
}

func SplitByStatus(devices []domain.Device) map[domain.DeviceStatus][]domain.Device {
	result := make(map[domain.DeviceStatus][]domain.Device)
	for _, device := range devices {
		result[device.Status] = append(result[device.Status], device)
	}
	for status := range result {
		result[status] = domain.SortDevices(result[status])
	}
	return result
}
