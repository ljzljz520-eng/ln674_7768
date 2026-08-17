package domain

import "time"

type AuditAction string

const (
	ActionSyncStarted   AuditAction = "sync_started"
	ActionSyncCompleted AuditAction = "sync_completed"
	ActionSyncFailed    AuditAction = "sync_failed"
	ActionConfirmed     AuditAction = "confirmed"
	ActionRetryQueued   AuditAction = "retry_queued"
)

type AuditEvent struct {
	ID        string            `json:"id"`
	RecordID  string            `json:"record_id"`
	Operator  string            `json:"operator"`
	Action    AuditAction       `json:"action"`
	Message   string            `json:"message"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func (e AuditEvent) IsConfirmation() bool {
	return e.Action == ActionConfirmed
}

func (e AuditEvent) Subject() string {
	if e.RecordID == "" {
		return "system"
	}
	return e.RecordID
}

func NewAuditEvent(id, recordID, operator string, action AuditAction, message string, at time.Time) AuditEvent {
	return AuditEvent{ID: id, RecordID: recordID, Operator: operator, Action: action, Message: message, CreatedAt: at, Metadata: map[string]string{}}
}
