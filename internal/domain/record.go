package domain

import "time"

type SyncState string

const (
	SyncPending SyncState = "pending"
	SyncRunning SyncState = "running"
	SyncSuccess SyncState = "success"
	SyncFailed  SyncState = "failed"
)

type SyncRecord struct {
	ID            string    `json:"id"`
	Vendor        string    `json:"vendor"`
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    time.Time `json:"finished_at"`
	State         SyncState `json:"state"`
	DeviceCount   int       `json:"device_count"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	Confirmations int       `json:"confirmations"`
	AuditEventIDs []string  `json:"audit_event_ids"`
	Revision      int       `json:"revision"`
}

func (r SyncRecord) IsFinished() bool {
	return r.State == SyncSuccess || r.State == SyncFailed
}

func (r SyncRecord) CanConfirm() bool {
	return r.State == SyncSuccess
}

func (r *SyncRecord) MarkStarted(at time.Time) {
	r.State = SyncRunning
	r.StartedAt = at
	r.FinishedAt = time.Time{}
	r.ErrorMessage = ""
	r.Revision++
}

func (r *SyncRecord) MarkSuccess(at time.Time, count int) {
	r.State = SyncSuccess
	r.FinishedAt = at
	r.DeviceCount = count
	r.ErrorMessage = ""
	r.Revision++
}

func (r *SyncRecord) MarkFailure(at time.Time, message string) {
	r.State = SyncFailed
	r.FinishedAt = at
	r.ErrorMessage = message
	r.Revision++
}

func (r *SyncRecord) AddAudit(id string) {
	for _, existing := range r.AuditEventIDs {
		if existing == id {
			return
		}
	}
	r.AuditEventIDs = append(r.AuditEventIDs, id)
}
