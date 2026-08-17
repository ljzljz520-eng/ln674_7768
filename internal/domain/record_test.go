package domain

import (
	"testing"
	"time"
)

func TestSyncRecordLifecycle(t *testing.T) {
	record := SyncRecord{ID: "r", State: SyncPending}
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	record.MarkStarted(start)
	if record.State != SyncRunning || record.IsFinished() {
		t.Fatalf("start state: %+v", record)
	}
	record.MarkSuccess(start.Add(time.Minute), 3)
	if !record.CanConfirm() || !record.IsFinished() || record.DeviceCount != 3 {
		t.Fatalf("success state: %+v", record)
	}
	record.AddAudit("a")
	record.AddAudit("a")
	if len(record.AuditEventIDs) != 1 {
		t.Fatalf("duplicate audit: %#v", record.AuditEventIDs)
	}
}

func TestRetryTaskSchedule(t *testing.T) {
	task := RetryTask{MaxAttempts: 2}
	now := time.Unix(10, 0)
	task.Schedule(now, "down")
	if !task.IsDue(now.Add(time.Minute)) || task.Attempt != 1 {
		t.Fatalf("task not due: %+v", task)
	}
	task.Schedule(now, "still down")
	if !task.Resolved {
		t.Fatal("task should be resolved after max attempts")
	}
}
