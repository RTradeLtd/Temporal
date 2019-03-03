package cli

import (
	"context"
	"os"
	"testing"
	"time"

	"gx/ipfs/QmQkW9fnCsg9SLHdViiAh6qfBppodsPZVpU92dZLqYtEfs/go-ipfs-cmds"
)

var root = &cmds.Command{
	Subcommands: map[string]*cmds.Command{
		"test": &cmds.Command{
			Run: func(req *cmds.Request, re cmds.ResponseEmitter, e cmds.Environment) error {
				err := cmds.EmitOnce(re, 42)

				time.Sleep(2 * time.Second)

				e.(env).ch <- struct{}{}
				return err
			},
		},
	},
}

type env struct {
	ch chan struct{}
}

func (e env) Context() context.Context {
	return context.Background()
}

func TestRunWaits(t *testing.T) {
	flag := make(chan struct{}, 1)

	devnull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer devnull.Close()

	err = Run(
		context.Background(),
		root,
		[]string{"test", "test"},
		devnull, devnull, devnull,
		func(ctx context.Context, req *cmds.Request) (cmds.Environment, error) {
			return env{flag}, nil
		},
		func(req *cmds.Request, env interface{}) (cmds.Executor, error) {
			return cmds.NewExecutor(req.Root), nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-flag:
	default:
		t.Fatal("expected flag to be raised")
	}
}
