package log

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.bobheadxi.dev/res"
	"go.bobheadxi.dev/zapx/ztest"
)

func Test_loggerMiddleware(t *testing.T) {
	type args struct {
		method      string
		path        string
		body        io.Reader
		middlewares []func(http.Handler) http.Handler
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"GET with requestID",
			args{"GET", "/", nil, []func(http.Handler) http.Handler{middleware.RequestID}},
			[]string{"path", "request-id"},
		},
		{
			"GET with realIP",
			args{"GET", "/", nil, []func(http.Handler) http.Handler{middleware.RealIP}},
			[]string{"path", "real-ip"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create bootstrapped logger and middleware
			var l, out = ztest.NewObservable()
			var handler = NewMiddleware(l.Sugar())

			// set up mock router
			m := chi.NewRouter()
			m.Use(tt.args.middlewares...)
			m.Use(handler)
			m.Get("/", func(w http.ResponseWriter, r *http.Request) {
				res.R(w, r, res.MsgOK("hello world!"))
			})

			// create a mock request to use
			req := httptest.NewRequest(tt.args.method, "http://testing"+tt.args.path,
				tt.args.body)

			// serve request
			m.ServeHTTP(httptest.NewRecorder(), req)

			// check for desired log fields
			for _, e := range out.All() {
				for _, f := range tt.want {
					// find field, cast as string, and check if empty
					if val, _ := e.ContextMap()[f].(string); val == "" {
						t.Errorf("field %s unexpectedly empty", f)
					}
				}
			}
		})
	}
}
