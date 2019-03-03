package http

import (
	"gx/ipfs/QmPVkJMTeRC6iBByPWdrRkD3BE5UXsj5HPzb4kPqL186mS/testify/mock"
	"net/http"
)

// TestRoundTripper DEPRECATED USE net/http/httptest
type TestRoundTripper struct {
	mock.Mock
}

// RoundTrip DEPRECATED USE net/http/httptest
func (t *TestRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := t.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
