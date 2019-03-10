package utils

import (
	"errors"
	"io"

	clamd "github.com/baruwa-enterprise/clamd"
)

// Shell is used to interact with clamav
type Shell struct {
	clam *clamd.Client
}

// NewShell is used to instantiate our connection to a clamav daemon
func NewShell(address string) (*Shell, error) {
	c, err := clamd.NewClient("tcp", address)
	if err != nil {
		return nil, err
	}
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
