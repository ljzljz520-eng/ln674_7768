package dashboard

import (
	"fmt"
	"sort"
	"strings"

	"campus-device-hub/internal/domain"
)

func SortCards(cards []VendorCard) []VendorCard {
	result := append([]VendorCard(nil), cards...)
	sort.Slice(result, func(i, j int) bool { return result[i].Vendor < result[j].Vendor })
	return result
}

func RenderSummary(summary domain.DashboardSummary) string {
	return fmt.Sprintf("devices=%d healthy=%d warning=%d offline=%d syncs=%d failed=%d retries=%d", summary.TotalDevices, summary.HealthyDevices, summary.WarningDevices, summary.OfflineDevices, summary.TotalSyncs, summary.FailedSyncs, summary.PendingRetries)
}

func RenderCards(cards []VendorCard) string {
	lines := make([]string, 0, len(cards))
	for _, card := range SortCards(cards) {
		lines = append(lines, fmt.Sprintf("%s %s %d/%d", card.Vendor, card.State, card.HealthyCount, card.DeviceCount))
	}
	return strings.Join(lines, "\n")
}
