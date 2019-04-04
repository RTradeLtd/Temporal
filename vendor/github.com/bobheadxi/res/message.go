package res

import (
	"net/http"

	"github.com/bobheadxi/res/base"
)

// MsgResponse is the template for a typical HTTP response for messages
type MsgResponse struct {
	*base.Response
}

// Msg is a shortcut for non-error statuses
func Msg(message string, code int, kvs ...interface{}) *MsgResponse {
	return &MsgResponse{base.NewResponse(message, code, kvs)}
}

// MsgOK is a shortcut for an ok-status response
func MsgOK(message string, kvs ...interface{}) *MsgResponse {
	return &MsgResponse{base.NewResponse(message, http.StatusOK, kvs)}
}
