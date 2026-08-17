package service

import (
	"sort"

	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

type ReportService struct {
	store *storage.Store
}

func NewReportService(store *storage.Store) *ReportService {
	return &ReportService{store: store}
}

func (r *ReportService) Summary() (domain.DashboardSummary, error) {
	devices, err := r.store.ListDevices()
	if err != nil {
		return domain.DashboardSummary{}, err
	}
	records, err := r.store.ListSyncRecords()
	if err != nil {
		return domain.DashboardSummary{}, err
	}
	tasks, err := r.store.ListRetryTasks()
	if err != nil {
		return domain.DashboardSummary{}, err
	}
	summary := domain.DashboardSummary{TotalDevices: len(devices), TotalSyncs: len(records)}
	for _, device := range devices {
		switch device.Status {
		case domain.StatusOnline:
			summary.HealthyDevices++
		case domain.StatusWarning:
			summary.WarningDevices++
		case domain.StatusOffline:
			summary.OfflineDevices++
		default:
			summary.UnknownDevices++
		}
	}
	for _, record := range records {
		if record.State == domain.SyncSuccess {
			summary.SuccessfulSyncs++
		}
		if record.State == domain.SyncFailed {
			summary.FailedSyncs++
		}
	}
	for _, task := range tasks {
		if !task.Resolved {
			summary.PendingRetries++
		}
	}
	return summary, nil
}

func (r *ReportService) VendorStates() (map[string]domain.SyncState, error) {
	records, err := r.store.ListSyncRecords()
	if err != nil {
		return nil, err
	}
	result := make(map[string]domain.SyncState, len(records))
	for _, record := range records {
		result[record.Vendor] = record.State
	}
	return result, nil
}

func (r *ReportService) SortedVendors() ([]string, error) {
	states, err := r.VendorStates()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(states))
	for name := range states {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
