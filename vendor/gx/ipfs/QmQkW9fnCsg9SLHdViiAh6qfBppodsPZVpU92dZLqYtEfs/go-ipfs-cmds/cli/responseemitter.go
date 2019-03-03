package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"gx/ipfs/QmQkW9fnCsg9SLHdViiAh6qfBppodsPZVpU92dZLqYtEfs/go-ipfs-cmds"
)

var _ ResponseEmitter = &responseEmitter{}

// NewResponseEmitter constructs a new response emitter that writes results to
// the console.
func NewResponseEmitter(stdout, stderr io.Writer, req *cmds.Request) (ResponseEmitter, error) {
	encType, enc, err := cmds.GetEncoder(req, stdout, cmds.TextNewline)

	return &responseEmitter{
		stdout:  stdout,
		stderr:  stderr,
		encType: encType,
		enc:     enc,
	}, err
}

// ResponseEmitter extends cmds.ResponseEmitter to give better control over the command line
type ResponseEmitter interface {
	cmds.ResponseEmitter

	Stdout() io.Writer
	Stderr() io.Writer

	// SetStatus sets the exit status for this command.
	SetStatus(int)

	// Status returns the exit status for the command.
	Status() int
}

type responseEmitter struct {
	l      sync.Mutex
	stdout io.Writer
	stderr io.Writer

	length  uint64
	enc     cmds.Encoder
	encType cmds.EncodingType
	exit    int
	closed  bool
}

func (re *responseEmitter) Type() cmds.PostRunType {
	return cmds.CLI
}

func (re *responseEmitter) SetLength(l uint64) {
	re.length = l
}

func (re *responseEmitter) isClosed() bool {
	re.l.Lock()
	defer re.l.Unlock()

	return re.closed
}

func (re *responseEmitter) Close() error {
	return re.CloseWithError(nil)
}

func (re *responseEmitter) CloseWithError(err error) error {
	re.l.Lock()
	defer re.l.Unlock()

	if re.closed {
		return cmds.ErrClosingClosedEmitter
	}
	re.closed = true

	var msg string
	if err != nil {
		if re.exit == 0 {
			// Default "error" exit code.
			re.exit = 1
		}
		switch err {
		case context.Canceled:
			msg = "canceled"
		case context.DeadlineExceeded:
			msg = "timed out"
		default:
			msg = err.Error()
		}

		fmt.Fprintln(re.stderr, "Error:", msg)
	}

	defer func() {
		re.stdout = nil
		re.stderr = nil
	}()

	var errStderr, errStdout error
	if f, ok := re.stderr.(*os.File); ok {
		errStderr = f.Sync()
	}
	if f, ok := re.stdout.(*os.File); ok {
		errStdout = f.Sync()
	}

	// ignore error if the operating system doesn't support syncing std{out,err}
	if errStderr != nil && !isSyncNotSupportedErr(errStderr) {
		return errStderr
	}
	if errStdout != nil && !isSyncNotSupportedErr(errStdout) {
		return errStdout
	}
	return nil
}

func (re *responseEmitter) Emit(v interface{}) error {
	var isSingle bool
	// unwrap
	if val, ok := v.(cmds.Single); ok {
		v = val.Value
		isSingle = true
	}

	// channel emission iteration
	if ch, ok := v.(chan interface{}); ok {
		v = (<-chan interface{})(ch)
	}
	if ch, isChan := v.(<-chan interface{}); isChan {
		return cmds.EmitChan(re, ch)
	}

	// TODO find a better solution for this.
	// Idea: use the actual cmd.Type and not *cmd.Type
	// would need to fix all commands though
	switch c := v.(type) {
	case *string:
		v = *c
	case *int:
		v = *c
	}

	if re.isClosed() {
		return cmds.ErrClosedEmitter
	}

	var err error

	switch t := v.(type) {
	case io.Reader:
		_, err = io.Copy(re.stdout, t)
		if err != nil {
			return err
		}
	default:
		if re.enc != nil {
			err = re.enc.Encode(v)
		} else {
			_, err = fmt.Fprintln(re.stdout, t)
		}
	}

	if isSingle {
		return re.CloseWithError(err)
	}

	return err
}

// Stderr returns the ResponseWriter's stderr
func (re *responseEmitter) Stderr() io.Writer {
	return re.stderr
}

// Stdout returns the ResponseWriter's stdout
func (re *responseEmitter) Stdout() io.Writer {
	return re.stdout
}

// SetStatus sets the exit status of the command.
func (re *responseEmitter) SetStatus(code int) {
	re.l.Lock()
	defer re.l.Unlock()
	re.exit = code
}

// Status _returns_ the exit status of the command.
func (re *responseEmitter) Status() int {
	re.l.Lock()
	defer re.l.Unlock()
	return re.exit
}
