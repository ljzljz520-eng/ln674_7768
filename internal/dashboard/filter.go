package dashboard

import (
	"sort"
	"strings"
	"time"

	"campus-device-hub/internal/domain"
)

type DeviceRow struct {
	ID      string
	Vendor  string
	Kind    domain.DeviceKind
	Label   string
	Status  string
	Updated time.Time
	Age     time.Duration
	Healthy bool
}

func (v *View) Rows(now time.Time) ([]DeviceRow, error) {
	devices, err := v.store.ListDevices()
	if err != nil {
		return nil, err
	}
	rows := make([]DeviceRow, 0, len(devices))
	for _, device := range devices {
		rows = append(rows, DeviceRow{ID: device.ID, Vendor: device.Vendor, Kind: device.Kind, Label: device.Label, Status: device.DisplayStatus(), Updated: device.UpdatedAt, Age: domain.AgeAt(device, now), Healthy: device.IsHealthy()})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows, nil
}

func FilterRows(rows []DeviceRow, text string, healthyOnly bool) []DeviceRow {
	needle := strings.ToLower(strings.TrimSpace(text))
	result := make([]DeviceRow, 0, len(rows))
	for _, row := range rows {
		if healthyOnly && !row.Healthy {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(row.Label+" "+row.Vendor+" "+row.ID), needle) {
			continue
		}
		result = append(result, row)
	}
	return result
}

func GroupRowsByVendor(rows []DeviceRow) map[string][]DeviceRow {
	result := make(map[string][]DeviceRow)
	for _, row := range rows {
		result[row.Vendor] = append(result[row.Vendor], row)
	}
	for vendor := range result {
		sort.Slice(result[vendor], func(i, j int) bool { return result[vendor][i].ID < result[vendor][j].ID })
	}
	return result
}

func StatusCounts(rows []DeviceRow) map[string]int {
	counts := make(map[string]int)
	for _, row := range rows {
		counts[row.Status]++
	}
	return counts
}

func RenderRows(rows []DeviceRow) string {
	parts := make([]string, 0, len(rows))
	for _, row := range rows {
		parts = append(parts, row.ID+" "+row.Status)
	}
	return strings.Join(parts, "\n")
}
