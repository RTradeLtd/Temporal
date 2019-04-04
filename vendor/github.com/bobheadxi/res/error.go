package res

import (
	"net/http"

	"github.com/bobheadxi/res/base"
)

// ErrResponse is the template for a typical HTTP response for errors
type ErrResponse struct {
	*base.Response
}

// Err is a basic error response constructor
func Err(message string, code int, kvs ...interface{}) *ErrResponse {
	return &ErrResponse{base.NewResponse(message, code, kvs)}
}

// ErrInternalServer is a shortcut for internal server errors. It should be
// accompanied by an actual error.
func ErrInternalServer(message string, err error, kvs ...interface{}) *ErrResponse {
	var b = base.NewResponse(message, http.StatusInternalServerError, kvs)
	b.Err = err.Error()
	return &ErrResponse{b}
}

// ErrBadRequest is a shortcut for bad requests
func ErrBadRequest(message string, kvs ...interface{}) *ErrResponse {
	return &ErrResponse{base.NewResponse(message, http.StatusBadRequest, kvs)}
}

// ErrUnauthorized is a shortcut for unauthorized requests
func ErrUnauthorized(message string, kvs ...interface{}) *ErrResponse {
	return &ErrResponse{base.NewResponse(message, http.StatusUnauthorized, kvs)}
}

// ErrForbidden is a shortcut for forbidden requests
func ErrForbidden(message string, kvs ...interface{}) *ErrResponse {
	return &ErrResponse{base.NewResponse(message, http.StatusForbidden, kvs)}
}

// ErrNotFound is a shortcut for forbidden requests
func ErrNotFound(message string, kvs ...interface{}) *ErrResponse {
	return &ErrResponse{base.NewResponse(message, http.StatusNotFound, kvs)}
}
