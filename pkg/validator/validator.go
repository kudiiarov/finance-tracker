// Package validator provides common validation utilities.
package validator

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Common validation errors.
var (
	ErrInvalidUUID = errors.New("invalid UUID format")
	ErrInvalidDate = errors.New("invalid date format")
	ErrEmptyValue  = errors.New("value cannot be empty")
)

// UUID validates a UUID string.
func UUID(s string) error {
	if _, err := uuid.Parse(s); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidUUID, s)
	}
	return nil
}

// Date validates a date string in YYYY-MM-DD format.
func Date(s string) error {
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return fmt.Errorf("%w: expected YYYY-MM-DD, got %s", ErrInvalidDate, s)
	}
	return nil
}

// NotEmpty validates that a string is not empty.
func NotEmpty(s string) error {
	if s == "" {
		return ErrEmptyValue
	}
	return nil
}
