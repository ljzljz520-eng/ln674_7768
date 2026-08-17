package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ValidationIssue struct {
	Field   string
	Message string
}

func (i ValidationIssue) Error() string {
	return i.Field + ": " + i.Message
}

func ValidateDevice(device Device) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if strings.TrimSpace(device.ID) == "" {
		issues = append(issues, ValidationIssue{"id", "device id is required"})
	}
	if device.Kind != KindAccess && device.Kind != KindBroadcast && device.Kind != KindLighting {
		issues = append(issues, ValidationIssue{"kind", "unsupported device kind"})
	}
	if strings.TrimSpace(device.Vendor) == "" {
		issues = append(issues, ValidationIssue{"vendor", "vendor is required"})
	}
	if strings.TrimSpace(device.Label) == "" {
		issues = append(issues, ValidationIssue{"label", "label is required"})
	}
	if device.Status == StatusUnknown {
		issues = append(issues, ValidationIssue{"status", "status must be observed"})
	}
	if device.UpdatedAt.IsZero() {
		issues = append(issues, ValidationIssue{"updated_at", "update time is required"})
	}
	return issues
}

func ValidateRecord(record SyncRecord) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if record.ID == "" {
		issues = append(issues, ValidationIssue{"id", "record id is required"})
	}
	if record.Vendor == "" {
		issues = append(issues, ValidationIssue{"vendor", "vendor is required"})
	}
	if record.State == SyncPending && !record.StartedAt.IsZero() {
		issues = append(issues, ValidationIssue{"state", "pending record cannot have a start time"})
	}
	if record.Confirmations < 0 {
		issues = append(issues, ValidationIssue{"confirmations", "confirmation count cannot be negative"})
	}
	if record.IsFinished() && record.FinishedAt.IsZero() {
		issues = append(issues, ValidationIssue{"finished_at", "finished record needs finish time"})
	}
	return issues
}

func ValidateAuditEvent(event AuditEvent) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if event.ID == "" {
		issues = append(issues, ValidationIssue{"id", "audit id is required"})
	}
	if event.RecordID == "" {
		issues = append(issues, ValidationIssue{"record_id", "record id is required"})
	}
	if event.Operator == "" {
		issues = append(issues, ValidationIssue{"operator", "operator is required"})
	}
	if event.Action == "" {
		issues = append(issues, ValidationIssue{"action", "action is required"})
	}
	if event.CreatedAt.IsZero() {
		issues = append(issues, ValidationIssue{"created_at", "event time is required"})
	}
	return issues
}

func ValidateRetryTask(task RetryTask) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	if task.ID == "" {
		issues = append(issues, ValidationIssue{"id", "retry id is required"})
	}
	if task.RecordID == "" {
		issues = append(issues, ValidationIssue{"record_id", "record id is required"})
	}
	if task.MaxAttempts < 1 {
		issues = append(issues, ValidationIssue{"max_attempts", "max attempts must be positive"})
	}
	if task.Attempt < 0 {
		issues = append(issues, ValidationIssue{"attempt", "attempt cannot be negative"})
	}
	return issues
}

func HasValidationIssues(issues []ValidationIssue) bool {
	return len(issues) > 0
}

func FormatIssues(issues []ValidationIssue) string {
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		parts = append(parts, issue.Error())
	}
	sort.Strings(parts)
	return strings.Join(parts, "; ")
}

func StatusFromSignal(signal string) DeviceStatus {
	switch strings.ToLower(strings.TrimSpace(signal)) {
	case "online", "ready", "ok", "up":
		return StatusOnline
	case "warning", "muted", "degraded":
		return StatusWarning
	case "offline", "blocked", "down", "fault":
		return StatusOffline
	default:
		return StatusUnknown
	}
}

func AgeAt(device Device, now time.Time) time.Duration {
	if device.UpdatedAt.IsZero() || now.Before(device.UpdatedAt) {
		return 0
	}
	return now.Sub(device.UpdatedAt)
}

func DescribeDevice(device Device) string {
	return fmt.Sprintf("%s [%s] %s at %s", device.Label, device.Vendor, device.DisplayStatus(), device.UpdatedAt.UTC().Format(time.RFC3339))
}
