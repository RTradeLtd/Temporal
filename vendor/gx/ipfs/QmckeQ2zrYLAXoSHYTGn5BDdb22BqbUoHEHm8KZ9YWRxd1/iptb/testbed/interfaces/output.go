package testbedi

import (
	"io"
)

// Output manages running, inprocess, a process
type Output interface {
	// Args is the cleaned up version of the input.
	Args() []string

	// Error is the error returned from the command, after it exited.
	Error() error

	// Code is the unix style exit code, set after the command exited.
	ExitCode() int

	Stdout() io.ReadCloser
	Stderr() io.ReadCloser
}
