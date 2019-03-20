package v2

import (
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
	//	var interfaceAPIResp interfaceAPIResponse
	req, err := http.NewRequest(
		"POST",
		"/v2/proxy/api/v0/pin/add?arg=QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv&stream-channels=true",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	writer := newCloseNotifyingRecorder()
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(writer, req)
}
