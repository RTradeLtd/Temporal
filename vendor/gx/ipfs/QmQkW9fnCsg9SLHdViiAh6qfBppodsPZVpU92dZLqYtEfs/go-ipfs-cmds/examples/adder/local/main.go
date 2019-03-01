package main

import (
	"context"
	"fmt"
	"os"

	"gx/ipfs/QmQkW9fnCsg9SLHdViiAh6qfBppodsPZVpU92dZLqYtEfs/go-ipfs-cmds/examples/adder"

	"gx/ipfs/QmQkW9fnCsg9SLHdViiAh6qfBppodsPZVpU92dZLqYtEfs/go-ipfs-cmds"
	"gx/ipfs/QmQkW9fnCsg9SLHdViiAh6qfBppodsPZVpU92dZLqYtEfs/go-ipfs-cmds/cli"
)

func main() {
	// parse the command path, arguments and options from the command line
	req, err := cli.Parse(context.TODO(), os.Args[1:], os.Stdin, adder.RootCmd)
	if err != nil {
		panic(err)
	}

	req.Options["encoding"] = cmds.Text

	// create an emitter
	cliRe, err := cli.NewResponseEmitter(os.Stdout, os.Stderr, req)
	if err != nil {
		panic(err)
	}

	wait := make(chan struct{})
	var re cmds.ResponseEmitter = cliRe
	if pr, ok := req.Command.PostRun[cmds.CLI]; ok {
		var (
			res   cmds.Response
			lower = re
		)

		re, res = cmds.NewChanResponsePair(req)

		go func() {
			defer close(wait)
			err := pr(res, lower)
			if err != nil {
				fmt.Println("error: ", err)
			}
		}()
	} else {
		close(wait)
	}

	adder.RootCmd.Call(req, re, nil)
	<-wait

	os.Exit(cliRe.Status())
}
