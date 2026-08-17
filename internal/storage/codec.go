package storage

import (
	"encoding/json"
	"fmt"
)

func encodeValue(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal value: %w", err)
	}
	return data, nil
}

func decodeValue(data []byte, target any) error {
	if len(data) == 0 {
		return ErrNotFound
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshal value: %w", err)
	}
	return nil
}

func cloneBytes(data []byte) []byte {
	return append([]byte(nil), data...)
}
