package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"campus-device-hub/internal/domain"
)

type Command struct {
	Name     string
	Vendor   string
	Operator string
	RecordID string
}

type CommandResult struct {
	Message string
	Record  domain.SyncRecord
	Event   domain.AuditEvent
}

func ParseCommand(line string) (Command, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return Command{}, errors.New("command is empty")
	}
	command := Command{Name: strings.ToLower(parts[0])}
	for _, part := range parts[1:] {
		key, value, found := strings.Cut(part, "=")
		if !found || value == "" {
			return Command{}, fmt.Errorf("invalid argument %q", part)
		}
		switch key {
		case "vendor":
			command.Vendor = value
		case "operator":
			command.Operator = value
		case "record":
			command.RecordID = value
		default:
			return Command{}, fmt.Errorf("unknown argument %q", key)
		}
	}
	return command, nil
}

func (s *SyncService) Execute(ctx context.Context, command Command) (CommandResult, error) {
	switch command.Name {
	case "sync":
		record, err := s.SyncVendor(ctx, command.Vendor)
		return CommandResult{Message: "sync completed", Record: record}, err
	case "reconcile":
		record, err := s.Reconcile(ctx, command.Vendor)
		return CommandResult{Message: "reconcile completed", Record: record}, err
	default:
		return CommandResult{}, fmt.Errorf("unsupported command %q", command.Name)
	}
}

func (s *SyncService) ConfirmCommand(command Command) (CommandResult, error) {
	if command.RecordID == "" || command.Operator == "" {
		return CommandResult{}, errors.New("record and operator are required")
	}
	confirmation := NewConfirmationService(s.store, s.clock, nil, nil)
	event, err := confirmation.Confirm(command.RecordID, command.Operator)
	return CommandResult{Message: "confirmation recorded", Event: event}, err
}

func (s *SyncService) DescribeRecord(record domain.SyncRecord) string {
	return fmt.Sprintf("%s vendor=%s state=%s devices=%d confirmations=%d", record.ID, record.Vendor, record.State, record.DeviceCount, record.Confirmations)
}
