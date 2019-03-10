package utils

import (
	"strings"
	"testing"
)

const (
	defaultURL = "127.0.0.1:3310"
	eicar      = `X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"`
)

func TestClam(t *testing.T) {
	t.Skip()
	s, err := NewShell(defaultURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Scan(strings.NewReader("hello")); err != nil {
		t.Fatal(err)
	}
	if err := s.Scan(strings.NewReader(eicar)); err == nil {
		t.Fatal(err)
	} else if err.Error() != "virus found" {
		t.Fatal("failed to get correct error message")
	}
	if err := s.Scan(strings.NewReader("HELLO" + eicar + "WORLD")); err == nil {
		t.Fatal(err)
	} else if err.Error() != "virus found" {
		t.Fatal("failed to get correct error message")
	}
}
