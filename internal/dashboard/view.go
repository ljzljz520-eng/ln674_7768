package dashboard

import (
	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/service"
	"campus-device-hub/internal/storage"
)

type View struct {
	report *service.ReportService
	store  *storage.Store
}

func NewView(store *storage.Store) *View {
	return &View{report: service.NewReportService(store), store: store}
}

func (v *View) Summary() (domain.DashboardSummary, error) {
	return v.report.Summary()
}

func (v *View) VendorCards() ([]VendorCard, error) {
	states, err := v.report.VendorStates()
	if err != nil {
		return nil, err
	}
	devices, err := v.store.ListDevices()
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	healthy := make(map[string]int)
	for _, device := range devices {
		counts[device.Vendor]++
		if device.IsHealthy() {
			healthy[device.Vendor]++
		}
	}
	cards := make([]VendorCard, 0, len(states))
	for vendor, state := range states {
		cards = append(cards, VendorCard{Vendor: vendor, State: state, DeviceCount: counts[vendor], HealthyCount: healthy[vendor]})
	}
	return SortCards(cards), nil
}

type VendorCard struct {
	Vendor       string
	State        domain.SyncState
	DeviceCount  int
	HealthyCount int
}

func (c VendorCard) HealthRatio() float64 {
	if c.DeviceCount == 0 {
		return 0
	}
	return float64(c.HealthyCount) / float64(c.DeviceCount)
}
