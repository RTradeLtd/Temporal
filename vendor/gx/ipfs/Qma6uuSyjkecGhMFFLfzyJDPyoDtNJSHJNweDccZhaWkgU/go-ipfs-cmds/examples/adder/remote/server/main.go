package main

import (
	"context"
	nethttp "net/http"

	"gx/ipfs/Qma6uuSyjkecGhMFFLfzyJDPyoDtNJSHJNweDccZhaWkgU/go-ipfs-cmds/examples/adder"

	http "gx/ipfs/Qma6uuSyjkecGhMFFLfzyJDPyoDtNJSHJNweDccZhaWkgU/go-ipfs-cmds/http"
)

type env struct{}

func (env) Context() context.Context {
	return context.TODO()
}

func main() {
	h := http.NewHandler(env{}, adder.RootCmd, http.NewServerConfig())

	// create http rpc server
	err := nethttp.ListenAndServe(":6798", h)
	if err != nil {
		panic(err)
	}
}
