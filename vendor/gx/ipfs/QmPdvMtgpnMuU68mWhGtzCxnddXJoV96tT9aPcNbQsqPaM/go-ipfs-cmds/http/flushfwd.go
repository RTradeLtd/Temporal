package http

import (
	"gx/ipfs/QmPdvMtgpnMuU68mWhGtzCxnddXJoV96tT9aPcNbQsqPaM/go-ipfs-cmds"
	"net/http"
)

type flushfwder struct {
	cmds.ResponseEmitter
	http.Flusher
}

func NewFlushForwarder(r cmds.ResponseEmitter, f http.Flusher) ResponseEmitter {
	return flushfwder{ResponseEmitter: r, Flusher: f}
}
