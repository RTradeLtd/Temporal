package commands

import (
	"fmt"
)

func NewUsageError(s string) error {
	return &UsageError{
		s,
	}
}

type UsageError struct {
	s string
}

func (e *UsageError) Error() string {
	return fmt.Sprintf("Usage Error: %s", e.s)
}
