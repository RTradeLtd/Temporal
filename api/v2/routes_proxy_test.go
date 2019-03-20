package v2

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
)

// see https://stackoverflow.com/questions/33968840/how-to-test-reverse-proxy-with-martini-in-go
type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func Test_API_Routes_Proxy(t *testing.T) {

	type args struct {
		method, call string
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
	}{
		{"Test-Success", args{
			"POST",
			"/v2/proxy/api/v0/pin/add?arg=QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv&stream-channels=true",
		}, 200},
		{"Test-Bad-Command", args{
			"POST",
			"/v2/proxy/api/v0/pin/rm?arg=QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv",
		}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// load configuration
			cfg, err := config.LoadConfig("../../testenv/config.json")
			if err != nil {
				t.Fatal(err)
			}
			db, err := loadDatabase(cfg)
			if err != nil {
				t.Fatal(err)
			}

			// setup fake mock clients
			fakeLens := &mocks.FakeLensV2Client{}
			fakeOrch := &mocks.FakeServiceClient{}
			fakeSigner := &mocks.FakeSignerClient{}

			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			if err := sendProxyRequest(
				api,
				tt.args.method,
				tt.args.call,
				tt.wantCode,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func sendProxyRequest(api *API, method, call string, wantCode int) error {
	req, err := http.NewRequest(method, call, nil)
	if err != nil {
		return err
	}
	writer := newCloseNotifyingRecorder()
	defer writer.close()
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(writer, req)
	if writer.Code != wantCode {
		return fmt.Errorf("wantCode = %v, response code = %v", wantCode, writer.Code)
	}
	return nil
}
