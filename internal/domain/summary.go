package domain

type DashboardSummary struct {
	TotalDevices    int
	HealthyDevices  int
	WarningDevices  int
	OfflineDevices  int
	UnknownDevices  int
	TotalSyncs      int
	SuccessfulSyncs int
	FailedSyncs     int
	PendingRetries  int
}

func (s DashboardSummary) HealthPercent() int {
	if s.TotalDevices == 0 {
		return 0
	}
	return s.HealthyDevices * 100 / s.TotalDevices
}

func (s DashboardSummary) HasFailures() bool {
	return s.FailedSyncs > 0 || s.PendingRetries > 0
}
