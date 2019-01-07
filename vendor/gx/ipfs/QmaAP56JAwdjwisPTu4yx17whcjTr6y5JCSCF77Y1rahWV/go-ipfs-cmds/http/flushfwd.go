package http

import (
	"gx/ipfs/QmaAP56JAwdjwisPTu4yx17whcjTr6y5JCSCF77Y1rahWV/go-ipfs-cmds"
	"net/http"
)

type flushfwder struct {
	cmds.ResponseEmitter
	http.Flusher
}

func NewFlushForwarder(r cmds.ResponseEmitter, f http.Flusher) ResponseEmitter {
	return flushfwder{ResponseEmitter: r, Flusher: f}
}
