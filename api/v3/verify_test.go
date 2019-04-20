package v3

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/api/v3/mocks"
	"github.com/RTradeLtd/database/models"
	"github.com/bobheadxi/res"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gopkg.in/dgrijalva/jwt-go.v3"
)

func TestAuthService_httpVerificationHandler(t *testing.T) {
	type args struct {
		r *http.Request
	}
	type mock struct {
		findUserErr      error
		validateEmailErr error
	}
	tests := []struct {
		name        string
		args        args
		mock        mock
		wantCode    int
		wantMessage string
		wantError   string
	}{
		{"no query params",
			args{httptest.NewRequest("GET", "https://bobheadxi.dev", nil)},
			mock{},
			http.StatusBadRequest,
			"user, token cannot be empty",
			""},
		{"bad token (not a JWT)",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", "not-a-token")
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{},
			http.StatusUnauthorized,
			"invalid token",
			"segments"},
		{"bad token (incorrect algorithm)",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := jwt.
					NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{}).
					SignedString([]byte(defaultJWT.Key))
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{},
			http.StatusUnauthorized,
			"invalid token",
			"expect hs512"},
		{"bad token (invalid signing key)",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := jwt.
					NewWithClaims(defaultJWT.SigningAlgo, jwt.MapClaims{}).
					SignedString([]byte("wrong!"))
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{},
			http.StatusUnauthorized,
			"invalid token",
			"signature is invalid"},
		{"bad token (expired)",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := jwt.
					NewWithClaims(defaultJWT.SigningAlgo, jwt.MapClaims{
						claimUser:   "bobheadxi",
						claimExpiry: time.Now().Add(-time.Minute).Unix(),
					}).
					SignedString([]byte(defaultJWT.Key))
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{},
			http.StatusUnauthorized,
			"invalid token",
			"expired"},
		{"wrong user in token",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := defaultValidToken()
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "postables")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{},
			http.StatusBadRequest,
			"user in token does not match",
			""},
		{"user not found",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := defaultValidToken()
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{
				findUserErr: errors.New("oh no"),
			},
			http.StatusNotFound,
			"user not found",
			""},
		{"no challenge",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := defaultValidToken()
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{},
			http.StatusBadRequest,
			"challenge in token is incorrect",
			""},
		{"fail to verify challenge",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := jwt.
					NewWithClaims(defaultJWT.SigningAlgo, jwt.MapClaims{
						claimUser:      "bobheadxi",
						claimChallenge: "lunch",
					}).
					SignedString([]byte(defaultJWT.Key))
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{
				validateEmailErr: errors.New("oh no"),
			},
			http.StatusInternalServerError,
			"unable to validate user",
			""},
		{"success",
			args{func() *http.Request {
				r := httptest.NewRequest("GET", "https://bobheadxi.dev", nil)
				token, err := jwt.
					NewWithClaims(defaultJWT.SigningAlgo, jwt.MapClaims{
						claimUser:      "bobheadxi",
						claimChallenge: "lunch",
					}).
					SignedString([]byte(defaultJWT.Key))
				require.NoError(t, err)
				q := r.URL.Query()
				q.Add("user", "bobheadxi")
				q.Add("token", token)
				r.URL.RawQuery = q.Encode()
				return r
			}()},
			mock{},
			http.StatusOK,
			"user verified",
			""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				users = &mocks.FakeUserManager{
					FindByUserNameStub: func(u string) (*models.User, error) {
						if tt.mock.findUserErr == nil {
							return &models.User{
								UserName:               "bobheadxi",
								EmailVerificationToken: "lunch",
							}, nil
						}
						return nil, tt.mock.findUserErr
					},
					ValidateEmailVerificationTokenStub: func(string, string) (*models.User, error) {
						return nil, tt.mock.validateEmailErr
					},
				}

				a = &AuthService{
					users:   users,
					usage:   nil,
					credits: nil,
					emails:  nil,
					jwt:     defaultJWT,
					dev:     true,
					l:       zaptest.NewLogger(t).Sugar(),
				}
			)

			w := httptest.NewRecorder()
			a.httpVerificationHandler(w, tt.args.r)
			resp, err := res.Unmarshal(w.Body)
			require.NoError(t, err)
			t.Logf("received response %#v", resp)
			assert.Equal(t, resp.HTTPStatusCode, tt.wantCode)
			assert.Contains(t, resp.Message, tt.wantMessage)
			assert.Contains(t, resp.Err, tt.wantError)
		})
	}
}
