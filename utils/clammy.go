package utils

import (
	"errors"
	"io"
	"time"

	clamd "github.com/baruwa-enterprise/clamd"
)

// Shell is used to interact with clamav
type Shell struct {
	clam *clamd.Client
}

// NewShell is used to instantiate our connection to a clamav daemon
func NewShell(address string) (*Shell, error) {
	if address == "" {
		address = "127.0.0.1:3310"
	}
	c, err := clamd.NewClient("tcp", address)
	if err != nil {
		return nil, err
	}
	// 3 connection retry attempts
	c.SetConnRetries(3)
	// sleep of 1.25 seconds
	c.SetConnSleep(1250 * time.Millisecond)
	// timeout of 10 minutes
	c.SetConnTimeout(10 * time.Minute)
	c.SetCmdTimeout(10 * time.Minute)

	return &Shell{
		clam: c,
	}, nil
}

// Scan is used to scan a file reader
func (s *Shell) Scan(reader io.Reader) error {
	resp, err := s.clam.ScanReader(reader)
	if err != nil {
		return err
	}
	for _, v := range resp {
		if v.Status == "FOUND" {
			return errors.New("virus found")
		}
	}
	return nil
}
