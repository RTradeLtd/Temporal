package iptbutil

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/ipfs/iptb/testbed/interfaces"
)

type Output struct {
	args []string

	exitcode int

	err    error
	stdout []byte
	stderr []byte
}

func NewOutput(args []string, stdout, stderr []byte, exitcode int, cmderr error) testbedi.Output {
	return &Output{
		args:     args,
		stdout:   stdout,
		stderr:   stderr,
		exitcode: exitcode,
		err:      cmderr,
	}
}

func (o *Output) Args() []string {
	return o.args
}

func (o *Output) Error() error {
	return o.err
}
func (o *Output) ExitCode() int {
	return o.exitcode
}

func (o *Output) Stdout() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(o.stdout))
}

func (o *Output) Stderr() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(o.stderr))
}
