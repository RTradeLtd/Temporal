package utils

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

const (
	defaultURL = "127.0.0.1:3310"
	eicarURL   = "https://gateway.temporal.cloud/ipfs/QmYwvxdRF92aW7rxB1xHE3g1rHqa88s9rCbeAhb2BNg4es"
	eicarFile  = "./eicar.txt"
)

func TestClam(t *testing.T) {
	s, err := NewShell(defaultURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := downloadEicar(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(eicarFile)
	eicarBytes, err := ioutil.ReadFile(eicarFile)
	if err != nil {
		t.Fatal(err)
	}
	eicar := string(eicarBytes)
	if err := s.Scan(strings.NewReader("hello")); err != nil {
		t.Fatal(err)
	}
	if err := s.Scan(strings.NewReader(eicar)); err == nil {
		t.Fatal(err)
	} else if err.Error() != "virus found" {
		t.Fatal("failed to get correct error message")
	}
	if err := s.Scan(strings.NewReader(eicar + "WORLD")); err == nil {
		t.Fatal("error expected")
	} else if err.Error() != "virus found" {
		t.Fatal("failed to get correct error message")
	}
}

func downloadEicar() error {
	resp, err := http.Get(eicarURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(eicarFile)
	if err != nil {
		return err
	}
	defer out.Close()
	if copied, err := io.Copy(out, resp.Body); err != nil {
		return err
	} else if copied == 0 {
		return errors.New("unknown copy error")
	}
	return nil
}
