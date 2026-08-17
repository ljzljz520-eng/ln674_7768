package service

import (
	"sort"
	"time"

	"campus-device-hub/internal/domain"
	"campus-device-hub/internal/storage"
)

type OperationsMetrics struct {
	DeviceTotal       int
	HealthyTotal      int
	AverageAge        time.Duration
	SyncTotal         int
	SuccessfulTotal   int
	FailedTotal       int
	ConfirmationTotal int
	AuditTotal        int
	RetryPending      int
	RetryResolved     int
}

func (s *SyncService) Metrics(now time.Time) (OperationsMetrics, error) {
	devices, err := s.store.ListDevices()
	if err != nil {
		return OperationsMetrics{}, err
	}
	records, err := s.store.ListSyncRecords()
	if err != nil {
		return OperationsMetrics{}, err
	}
	events, err := s.store.ListAuditEvents()
	if err != nil {
		return OperationsMetrics{}, err
	}
	tasks, err := s.store.ListRetryTasks()
	if err != nil {
		return OperationsMetrics{}, err
	}
	metrics := OperationsMetrics{DeviceTotal: len(devices), SyncTotal: len(records), AuditTotal: len(events)}
	var totalAge time.Duration
	for _, device := range devices {
		if device.IsHealthy() {
			metrics.HealthyTotal++
		}
		totalAge += domain.AgeAt(device, now)
	}
	if len(devices) > 0 {
		metrics.AverageAge = totalAge / time.Duration(len(devices))
	}
	for _, record := range records {
		if record.State == domain.SyncSuccess {
			metrics.SuccessfulTotal++
		}
		if record.State == domain.SyncFailed {
			metrics.FailedTotal++
		}
		metrics.ConfirmationTotal += record.Confirmations
	}
	for _, task := range tasks {
		if task.Resolved {
			metrics.RetryResolved++
		} else {
			metrics.RetryPending++
		}
	}
	return metrics, nil
}

func (m OperationsMetrics) AvailabilityPercent() int {
	if m.DeviceTotal == 0 {
		return 0
	}
	return m.HealthyTotal * 100 / m.DeviceTotal
}

func (m OperationsMetrics) SyncSuccessPercent() int {
	if m.SyncTotal == 0 {
		return 0
	}
	return m.SuccessfulTotal * 100 / m.SyncTotal
}

func (m OperationsMetrics) NeedsAttention() bool {
	return m.FailedTotal > 0 || m.RetryPending > 0 || m.AvailabilityPercent() < 100
}

type VendorMetric struct {
	Vendor       string
	DeviceCount  int
	HealthyCount int
	SyncCount    int
	FailureCount int
}

func (s *SyncService) VendorMetrics() ([]VendorMetric, error) {
	devices, err := s.store.ListDevices()
	if err != nil {
		return nil, err
	}
	records, err := s.store.ListSyncRecords()
	if err != nil {
		return nil, err
	}
	byVendor := make(map[string]*VendorMetric)
	for _, device := range devices {
		metric := byVendor[device.Vendor]
		if metric == nil {
			metric = &VendorMetric{Vendor: device.Vendor}
			byVendor[device.Vendor] = metric
		}
		metric.DeviceCount++
		if device.IsHealthy() {
			metric.HealthyCount++
		}
	}
	for _, record := range records {
		metric := byVendor[record.Vendor]
		if metric == nil {
			metric = &VendorMetric{Vendor: record.Vendor}
			byVendor[record.Vendor] = metric
		}
		metric.SyncCount++
		if record.State == domain.SyncFailed {
			metric.FailureCount++
		}
	}
	result := make([]VendorMetric, 0, len(byVendor))
	for _, metric := range byVendor {
		result = append(result, *metric)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Vendor < result[j].Vendor })
	return result, nil
}

func NewMetrics(store *storage.Store, clock Clock) *MetricsReader {
	return &MetricsReader{store: store, clock: clock}
}

type MetricsReader struct {
	store *storage.Store
	clock Clock
}

func (m *MetricsReader) Read() (OperationsMetrics, error) {
	service := &SyncService{store: m.store, clock: m.clock}
	return service.Metrics(m.clock.Now())
}
