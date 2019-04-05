package v3

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bobheadxi/res"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gopkg.in/dgrijalva/jwt-go.v3"

	"github.com/RTradeLtd/Temporal/api/v3/mocks"
	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/models"
)

var defaultJWT = JWTConfig{
	Key:         "hello-world",
	Timeout:     time.Minute,
	SigningAlgo: jwt.SigningMethodHS512,
}

func defaultValidToken() (string, error) {
	return jwt.
		NewWithClaims(defaultJWT.SigningAlgo, jwt.MapClaims{
			claimUser:   "bobheadxi",
			claimExpiry: time.Now().Add(10 * time.Minute).Unix(),
		}).
		SignedString([]byte(defaultJWT.Key))
}

func TestAuthService_Register(t *testing.T) {
	type args struct {
		ctx context.Context
		req *auth.RegisterReq
	}
	type mock struct {
		newUserErr  error
		genEmailErr error
		newUsageErr error
		publishErr  error
	}
	tests := []struct {
		name    string
		args    args
		mock    mock
		wantErr codes.Code
	}{
		{"email cannot have +",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bob+rob@bobheadxi.dev",
				Credentials:  &auth.Credentials{},
			}},
			mock{},
			codes.InvalidArgument},
		{"invalid credentials",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{},
			}},
			mock{},
			codes.InvalidArgument},
		{"duplicate email",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{Username: "bobheadxi", Password: "sekret"},
			}},
			mock{
				newUserErr: errors.New(eh.DuplicateEmailError),
			},
			codes.InvalidArgument},
		{"duplicate username",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{Username: "bobheadxi", Password: "sekret"},
			}},
			mock{
				newUserErr: errors.New(eh.DuplicateUserNameError),
			},
			codes.InvalidArgument},
		{"unexpected error",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{Username: "bobheadxi", Password: "sekret"},
			}},
			mock{
				newUserErr: errors.New("uh oh"),
			},
			codes.Internal},
		{"fail to generate email token",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{Username: "bobheadxi", Password: "sekret"},
			}},
			mock{
				genEmailErr: errors.New("uh oh"),
			},
			codes.Internal},
		{"fail to send email",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{Username: "bobheadxi", Password: "sekret"},
			}},
			mock{
				publishErr: errors.New("oh no"),
			},
			codes.Internal},
		{"fail to generate new usage entry",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{Username: "bobheadxi", Password: "sekret"},
			}},
			mock{
				newUsageErr: errors.New("oh no"),
			},
			codes.Internal},
		{"success",
			args{c(), &auth.RegisterReq{
				EmailAddress: "bobrob@bobheadxi.dev",
				Credentials:  &auth.Credentials{Username: "bobheadxi", Password: "sekret"},
			}},
			mock{},
			codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				users = &mocks.FakeUserManager{
					NewUserAccountStub: func(u string, p string, e string) (*models.User, error) {
						return nil, tt.mock.newUserErr
					},
					GenerateEmailVerificationTokenStub: func(string) (*models.User, error) {
						if tt.mock.genEmailErr == nil {
							return &models.User{
								UserName:     tt.args.req.GetCredentials().GetPassword(),
								EmailAddress: tt.args.req.GetEmailAddress(),
							}, nil
						}
						return nil, tt.mock.genEmailErr
					},
				}
				usage = &mocks.FakeUsageManager{
					NewUsageEntryStub: func(string, models.DataUsageTier) (*models.Usage, error) {
						if tt.mock.newUsageErr == nil {
							return &models.Usage{
								Tier: models.Free,
							}, nil
						}
						return nil, tt.mock.newUsageErr
					},
				}
				emails = &mocks.FakePublisher{
					PublishMessageStub: func(interface{}) error { return tt.mock.publishErr },
				}

				a = &AuthService{
					users:   users,
					usage:   usage,
					credits: nil,
					emails:  emails,
					jwt:     defaultJWT,
					dev:     true,
					l:       zaptest.NewLogger(t).Sugar(),
				}
			)

			got, err := a.Register(tt.args.ctx, tt.args.req)
			if !assert.Equal(t, tt.wantErr, status.Code(err)) {
				t.Logf("got error = %v", err)
			}
			if tt.wantErr == codes.OK {
				require.NotNil(t, got)
				assert.Equal(t, tt.args.req.GetEmailAddress(), got.GetEmailAddress())
				assert.False(t, got.GetVerified())
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	type args struct {
		ctx context.Context
		req *auth.Credentials
	}
	type mock struct {
		signInOk  bool
		signInErr error
	}
	tests := []struct {
		name    string
		args    args
		mock    mock
		wantErr codes.Code
	}{
		{"invalid arguments",
			args{c(), &auth.Credentials{}},
			mock{},
			codes.InvalidArgument},
		{"invalid credentials",
			args{c(), &auth.Credentials{Username: "bobheadxi", Password: "isthebest"}},
			mock{signInOk: false},
			codes.Unauthenticated},
		{"error when signing in",
			args{c(), &auth.Credentials{Username: "bobheadxi", Password: "isthebest"}},
			mock{signInErr: errors.New("oh no")},
			codes.Internal},
		{"success",
			args{c(), &auth.Credentials{Username: "bobheadxi", Password: "isthebest"}},
			mock{signInOk: true},
			codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				users = &mocks.FakeUserManager{
					SignInStub: func(u string, p string) (bool, error) {
						return tt.mock.signInOk, tt.mock.signInErr
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
			got, err := a.Login(tt.args.ctx, tt.args.req)
			if !assert.Equal(t, tt.wantErr, status.Code(err)) {
				t.Logf("got error = %v", err)
			}
			if tt.wantErr == codes.OK {
				require.NotNil(t, got)
				assert.True(t, time.Unix(got.GetExpire(), 0).After(time.Now()),
					"expiry should be after now")
				assert.NotEmpty(t, got.GetToken())
			}
		})
	}
}

func TestAuthService_Account(t *testing.T) {
	type args struct {
		ctx context.Context
		req *auth.Empty
	}
	type mock struct {
		findUserErr error
	}
	tests := []struct {
		name    string
		args    args
		mock    mock
		wantErr codes.Code
	}{
		{"no user in context",
			args{c(), &auth.Empty{}},
			mock{},
			codes.NotFound},
		{"error getting user usage",
			args{cFromMap(cMap{ctxKeyUser: &models.User{
				UserName: "bobheadxi",
			}}), &auth.Empty{}},
			mock{
				findUserErr: errors.New("oh no"),
			},
			codes.Internal},
		{"success",
			args{cFromMap(cMap{ctxKeyUser: &models.User{
				UserName: "bobheadxi",
			}}), &auth.Empty{}},
			mock{},
			codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				usage = &mocks.FakeUsageManager{
					FindByUserNameStub: func(string) (*models.Usage, error) {
						if tt.mock.findUserErr == nil {
							return &models.Usage{
								Tier: models.Partner,
							}, nil
						}
						return nil, tt.mock.findUserErr
					},
				}

				a = &AuthService{
					users:   nil,
					usage:   usage,
					credits: nil,
					emails:  nil,
					jwt:     defaultJWT,
					dev:     true,
					l:       zaptest.NewLogger(t).Sugar(),
				}
			)

			got, err := a.Account(tt.args.ctx, tt.args.req)
			if !assert.Equal(t, tt.wantErr, status.Code(err)) {
				t.Logf("got error = %v", err)
			}
			if tt.wantErr == codes.OK {
				require.NotNil(t, got)
				assert.Equal(t, auth.Tier_PARTNER, got.GetUsage().GetTier())
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	type args struct {
		ctx context.Context
		req *auth.Empty
	}
	type mock struct {
	}
	tests := []struct {
		name    string
		args    args
		mock    mock
		wantErr codes.Code
	}{
		{"no user in context",
			args{c(), &auth.Empty{}},
			mock{},
			codes.NotFound},
		{"success",
			args{cFromMap(cMap{ctxKeyUser: &models.User{
				UserName: "bobheadxi",
			}}), &auth.Empty{}},
			mock{},
			codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				a = &AuthService{
					users:   nil,
					usage:   nil,
					credits: nil,
					emails:  nil,
					jwt:     defaultJWT,
					dev:     true,
					l:       zaptest.NewLogger(t).Sugar(),
				}
			)

			got, err := a.Refresh(tt.args.ctx, tt.args.req)
			assert.Equalf(t, tt.wantErr, status.Code(err),
				"got { %v }", err)
			if tt.wantErr == codes.OK {
				require.NotNil(t, got)
				assert.True(t, time.Unix(got.GetExpire(), 0).After(time.Now()),
					"expiry should be after now")
				assert.NotEmpty(t, got.GetToken())
			}
		})
	}
}

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

func TestAuthService_newAuthInterceptors(t *testing.T) {
	const defaultMethod = "/auth.TemporalAuth/Register"
	token, err := defaultValidToken()
	require.NoError(t, err)

	type args struct {
		exceptions []string
		ctx        context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"context without key should be rejected",
			args{nil, c()},
			true},
		{"context without key should be allowed if in exceptions",
			args{[]string{defaultMethod}, c()},
			false},
		{"context with metadata and key should be allowed",
			args{nil, metadata.NewIncomingContext(c(), metadata.MD{
				"authorization": []string{"Bearer " + token},
			})},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthService{
				users: &mocks.FakeUserManager{
					FindByUserNameStub: func(u string) (*models.User, error) {
						return &models.User{
							UserName: "bobheadxi",
						}, nil
					},
				},
				usage:  nil,
				emails: nil,
				jwt:    defaultJWT,
				dev:    true,
				l:      zaptest.NewLogger(t).Sugar(),
			}
			unary, stream := a.newAuthInterceptors(tt.args.exceptions...)

			var called bool

			_, err := unary(tt.args.ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: defaultMethod,
			}, func(context.Context, interface{}) (interface{}, error) {
				called = true
				return nil, nil
			})
			if tt.wantErr {
				assert.Equalf(t, codes.Unauthenticated, status.Code(err),
					"got { %v }", err)
				assert.False(t, called, "handler should not have been called")
			} else {
				assert.Equalf(t, codes.OK, status.Code(err),
					"got { %v }", err)
				assert.True(t, called, "handler should have been called")
			}

			called = false
			err = stream(nil, &mocks.FakeServerStream{
				ContextStub: func() context.Context { return tt.args.ctx },
			}, &grpc.StreamServerInfo{
				FullMethod: defaultMethod,
			}, func(interface{}, grpc.ServerStream) error {
				called = true
				return nil
			})
			if tt.wantErr {
				assert.Equalf(t, codes.Unauthenticated, status.Code(err),
					"got { %v }", err)
				assert.False(t, called, "handler should not have been called")
			} else {
				assert.Equalf(t, codes.OK, status.Code(err),
					"got { %v }", err)
				assert.True(t, called, "handler should have been called")
			}
		})
	}
}

func TestAuthService_validate(t *testing.T) {
	validToken, err := defaultValidToken()
	require.NoError(t, err)

	type args struct {
		ctx context.Context
	}
	type mock struct {
		findUserErr error
	}
	tests := []struct {
		name    string
		args    args
		mock    mock
		wantErr bool
	}{
		{"no token", args{c()}, mock{}, true},
		{"expired token",
			args{metadata.NewIncomingContext(c(), metadata.MD{
				"authorization": []string{"Bearer " + func() string {
					expired, err := jwt.
						NewWithClaims(defaultJWT.SigningAlgo, jwt.MapClaims{
							claimUser:   "bobheadxi",
							claimExpiry: time.Now().Add(-time.Minute).Unix(),
						}).
						SignedString([]byte(defaultJWT.Key))
					require.NoError(t, err)
					return expired
				}()},
			})},
			mock{
				findUserErr: errors.New("oh no"),
			},
			true},
		{"unable to find user",
			args{metadata.NewIncomingContext(c(), metadata.MD{
				"authorization": []string{"Bearer " + validToken},
			})},
			mock{
				findUserErr: errors.New("oh no"),
			},
			true},
		{"success",
			args{metadata.NewIncomingContext(c(), metadata.MD{
				"authorization": []string{"Bearer " + validToken},
			})},
			mock{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				users = &mocks.FakeUserManager{
					FindByUserNameStub: func(u string) (*models.User, error) {
						if tt.mock.findUserErr == nil {
							return &models.User{
								UserName: "bobheadxi",
							}, nil
						}
						return nil, tt.mock.findUserErr
					},
				}

				a = &AuthService{
					users:  users,
					usage:  nil,
					emails: nil,
					jwt:    defaultJWT,
					dev:    true,
					l:      zaptest.NewLogger(t).Sugar(),
				}
			)

			got, err := a.validate(tt.args.ctx)
			if tt.wantErr {
				assert.Equalf(t, codes.Unauthenticated, status.Code(err),
					"got { %v }", err)
			} else {
				assert.Equalf(t, codes.OK, status.Code(err),
					"got { %v }", err)
				require.NotNil(t, got)
				u, b := ctxGetUser(got)
				assert.True(t, b)
				assert.NotNil(t, u)
				assert.Equal(t, "bobheadxi", u.UserName)
			}
		})
	}
}

func Test_toUser(t *testing.T) {
	type args struct {
		u     *models.User
		usage *models.Usage
	}
	tests := []struct {
		name string
		args args
		want *auth.User
	}{
		{"zero-value should be equal and not panic",
			args{&models.User{}, nil},
			&auth.User{
				IpfsKeys:  map[string]string{}, // always instantiated
				ApiAccess: true,                // hardcoded
				Usage:     nil,                 // always instantiated
			}},
		{"should iterate ipfs keys",
			args{&models.User{
				IPFSKeyIDs:   []string{"robert"},
				IPFSKeyNames: []string{"bobheadxi"},
			}, nil},
			&auth.User{
				IpfsKeys: map[string]string{
					"robert": "bobheadxi",
				},
				ApiAccess: true, // hardcoded
				Usage:     nil,
			}},
		{"should set tier partner",
			args{&models.User{}, &models.Usage{
				Tier: models.Partner,
			}},
			&auth.User{
				IpfsKeys:  map[string]string{}, // always instantiated
				ApiAccess: true,                // hardcoded
				Usage: &auth.User_Usage{
					Tier:        auth.Tier_PARTNER,
					Data:        &auth.User_Usage_Limits{},
					IpnsRecords: &auth.User_Usage_Limits{},
					PubsubSent:  &auth.User_Usage_Limits{},
					Keys:        &auth.User_Usage_Limits{},
				},
			}},
		{"should set tier light",
			args{&models.User{}, &models.Usage{
				Tier: models.Light,
			}},
			&auth.User{
				IpfsKeys:  map[string]string{}, // always instantiated
				ApiAccess: true,                // hardcoded
				Usage: &auth.User_Usage{
					Tier:        auth.Tier_LIGHT,
					Data:        &auth.User_Usage_Limits{},
					IpnsRecords: &auth.User_Usage_Limits{},
					PubsubSent:  &auth.User_Usage_Limits{},
					Keys:        &auth.User_Usage_Limits{},
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// hack around the fact the deep deep equality isn't currently implemented
			// as expected. possible resolution in https://github.com/stretchr/testify/issues/535
			u := toUser(tt.args.u, tt.args.usage)
			assert.Equal(t, tt.want.GetUsage().GetTier(), u.GetUsage().GetTier())
			assert.Equal(t, tt.want.GetUsage().GetData(), u.GetUsage().GetData())
			assert.Equal(t, tt.want.GetUsage().GetIpnsRecords(), u.GetUsage().GetIpnsRecords())
			assert.Equal(t, tt.want.GetUsage().GetPubsubSent(), u.GetUsage().GetPubsubSent())
			assert.Equal(t, tt.want.GetUsage().GetKeys(), u.GetUsage().GetKeys())
			u.Usage = nil
			tt.want.Usage = nil
			assert.Equal(t, tt.want, u)
		})
	}
}
