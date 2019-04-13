package v3

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/api/v3/mocks"
	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/models"
	"github.com/bobheadxi/res"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gopkg.in/dgrijalva/jwt-go.v3"
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

func TestAuthService_Recover(t *testing.T) {
	type args struct {
		ctx context.Context
		req *auth.RecoverReq
	}
	type mock struct {
		findByEmailErr    error
		emailVerified     bool
		resetPasswordErr  error
		publishMessageErr error
	}
	tests := []struct {
		name    string
		args    args
		mock    mock
		wantErr codes.Code
	}{
		{"no email", args{c(), &auth.RecoverReq{}}, mock{}, codes.InvalidArgument},
		{"invalid type",
			args{c(), &auth.RecoverReq{
				Type:         -1,
				EmailAddress: "robert@bobheadxi.dev",
			}}, mock{
				emailVerified: true,
			}, codes.InvalidArgument},
		{"no such user",
			args{c(), &auth.RecoverReq{
				EmailAddress: "robert@bobheadxi.dev",
			}},
			mock{
				findByEmailErr: errors.New("oh no"),
			},
			codes.NotFound},
		{"user email not enabled",
			args{c(), &auth.RecoverReq{
				EmailAddress: "robert@bobheadxi.dev",
			}},
			mock{
				emailVerified: false,
			},
			codes.FailedPrecondition},
		{"password: error when resetting",
			args{c(), &auth.RecoverReq{
				Type:         auth.RecoverReq_PASSWORD,
				EmailAddress: "robert@bobheadxi.dev",
			}},
			mock{
				emailVerified:    true,
				resetPasswordErr: errors.New("oh no"),
			},
			codes.Internal},
		{"password: error when publishing",
			args{c(), &auth.RecoverReq{
				Type:         auth.RecoverReq_PASSWORD,
				EmailAddress: "robert@bobheadxi.dev",
			}},
			mock{
				emailVerified:     true,
				publishMessageErr: errors.New("oh no"),
			},
			codes.Internal},
		{"password: success",
			args{c(), &auth.RecoverReq{
				Type:         auth.RecoverReq_PASSWORD,
				EmailAddress: "robert@bobheadxi.dev",
			}},
			mock{
				emailVerified: true,
			},
			codes.OK},
		{"username: error when publishing",
			args{c(), &auth.RecoverReq{
				Type:         auth.RecoverReq_USERNAME,
				EmailAddress: "robert@bobheadxi.dev",
			}},
			mock{
				emailVerified:     true,
				publishMessageErr: errors.New("oh no"),
			},
			codes.Internal},
		{"username: success",
			args{c(), &auth.RecoverReq{
				Type:         auth.RecoverReq_USERNAME,
				EmailAddress: "robert@bobheadxi.dev",
			}},
			mock{
				emailVerified: true,
			},
			codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				users = &mocks.FakeUserManager{
					FindByEmailStub: func(string) (*models.User, error) {
						if tt.mock.findByEmailErr == nil {
							return &models.User{
								UserName:     "bobheadxi",
								EmailAddress: "robert@bobheadxi.dev",
								EmailEnabled: tt.mock.emailVerified,
							}, nil
						}
						return nil, tt.mock.findByEmailErr
					},
					ResetPasswordStub: func(string) (string, error) {
						return "", tt.mock.resetPasswordErr
					},
				}
				emails = &mocks.FakePublisher{
					PublishMessageStub: func(interface{}) error {
						return tt.mock.publishMessageErr
					},
				}

				a = &AuthService{
					users:   users,
					usage:   nil,
					credits: nil,
					emails:  emails,
					jwt:     defaultJWT,
					dev:     true,
					l:       zaptest.NewLogger(t).Sugar(),
				}
			)

			got, err := a.Recover(tt.args.ctx, tt.args.req)
			assert.Equalf(t, tt.wantErr, status.Code(err), "got err = '%v'", err)
			if tt.wantErr == codes.OK {
				require.NotNil(t, got)
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
			assert.Equalf(t, tt.wantErr, status.Code(err), "got err = '%v'", err)
			if tt.wantErr == codes.OK {
				require.NotNil(t, got)
				assert.Equal(t, auth.Tier_PARTNER, got.GetUsage().GetTier())
			}
		})
	}
}

func TestAuthService_Update(t *testing.T) {
	type args struct {
		ctx context.Context
		req *auth.UpdateReq
	}
	type mock struct {
		changePasswordErr error
		changePasswordOK  bool

		findUsageErr  error
		findUsageTier models.DataUsageTier

		updateTierErr     error
		addCreditsErr     error
		publishMessageErr error
	}
	tests := []struct {
		name    string
		args    args
		mock    mock
		wantErr codes.Code
	}{
		{"no user in ctx", args{c(), &auth.UpdateReq{}}, mock{}, codes.NotFound},
		{"invalid type",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: nil,
			}},
			mock{},
			codes.InvalidArgument},
		{"password: incorrect password",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_PasswordChange{
					PasswordChange: &auth.UpdateReq_Password{
						OldPassword: "asdf", NewPassword: "ioiu",
					}},
			}},
			mock{
				changePasswordOK: false,
			},
			codes.PermissionDenied},
		{"password: error when changing",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_PasswordChange{
					PasswordChange: &auth.UpdateReq_Password{
						OldPassword: "asdf", NewPassword: "ioiu",
					}},
			}},
			mock{
				changePasswordErr: errors.New("oh no"),
			},
			codes.Internal},
		{"password: ok",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_PasswordChange{
					PasswordChange: &auth.UpdateReq_Password{
						OldPassword: "asdf", NewPassword: "ioiu",
					}},
			}},
			mock{
				changePasswordOK: true,
			},
			codes.OK},
		{"tier: cant find usage",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_DataTierChange{
					DataTierChange: &auth.UpdateReq_DataTier{}},
			}},
			mock{
				findUsageErr: errors.New("oh no"),
			},
			codes.Internal},
		{"tier: user is already upgraded",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_DataTierChange{
					DataTierChange: &auth.UpdateReq_DataTier{}},
			}},
			mock{
				findUsageTier: models.Light,
			},
			codes.AlreadyExists},
		{"tier: fail to add credits",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_DataTierChange{
					DataTierChange: &auth.UpdateReq_DataTier{}},
			}},
			mock{
				findUsageTier: models.Free,
				addCreditsErr: errors.New("oh no"),
			},
			codes.Internal},
		{"tier: fail to publish email to queue",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_DataTierChange{
					DataTierChange: &auth.UpdateReq_DataTier{}},
			}},
			mock{
				findUsageTier:     models.Free,
				publishMessageErr: errors.New("oh no"),
			},
			codes.Internal},
		{"tier: success",
			args{cFromMap(cMap{
				ctxKeyUser: &models.User{UserName: "bobheadxi"},
			}), &auth.UpdateReq{
				Update: &auth.UpdateReq_DataTierChange{
					DataTierChange: &auth.UpdateReq_DataTier{}},
			}},
			mock{
				findUsageTier: models.Free,
			},
			codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				users = &mocks.FakeUserManager{
					ChangePasswordStub: func(string, string, string) (bool, error) {
						return tt.mock.changePasswordOK, tt.mock.changePasswordErr
					},
				}
				usage = &mocks.FakeUsageManager{
					FindByUserNameStub: func(string) (*models.Usage, error) {
						if tt.mock.findUsageErr == nil {
							return &models.Usage{
								Tier: tt.mock.findUsageTier,
							}, nil
						}
						return nil, tt.mock.findUsageErr
					},
					UpdateTierStub: func(string, models.DataUsageTier) error {
						return tt.mock.updateTierErr
					},
				}
				credits = &mocks.FakeCreditsManager{
					AddCreditsStub: func(string, float64) (*models.User, error) {
						if tt.mock.addCreditsErr == nil {
							return &models.User{
								UserName: "bobheadxi",
							}, nil
						}
						return nil, tt.mock.addCreditsErr
					},
				}
				emails = &mocks.FakePublisher{
					PublishMessageStub: func(interface{}) error {
						return tt.mock.publishMessageErr
					},
				}

				a = &AuthService{
					users:   users,
					usage:   usage,
					credits: credits,
					emails:  emails,
					jwt:     defaultJWT,
					dev:     true,
					l:       zaptest.NewLogger(t).Sugar(),
				}
			)

			got, err := a.Update(tt.args.ctx, tt.args.req)
			assert.Equalf(t, tt.wantErr, status.Code(err), "got err = '%v'", err)
			if tt.wantErr == codes.OK {
				require.NotNil(t, got)
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

func Test_validateEmailFormat(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"not an email", args{"bobheadxi"}, true},
		{"no '+' in account", args{"rob+temporal@bobheadxi.dev"}, true},
		{"no trailing '.' in account", args{"rob.@bobheadxi.dev"}, true},
		{"no trailing '....' in account", args{"rob....@bobheadxi.dev"}, true},
		{"invalid email domain", args{"rob@hahahaha"}, true},
		{"valid custom email domain", args{"rob@bobheadxi.dev"}, false},
		{"valid gmail", args{"rob@gmail.com"}, false},
		{"valid hotmail", args{"rob@gmail.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateEmailFormat(tt.args.email); (err != nil) != tt.wantErr {
				t.Errorf("validateEmailFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	if os.Getenv("TEST_EMAIL_DOMAINS") == "true" {
		t.Run("all popular domains should validate", func(t *testing.T) {
			for _, domain := range commonEmailDomains {
				if err := validateEmailFormat("rob@" + domain); err != nil {
					t.Errorf("validateEmailFormat() error = %v", err)
				}
			}
		})
	}
}

// sourced from https://github.com/mailcheck/mailcheck/wiki/List-of-Popular-Domains
// a few domains were removed because they can't be reached:
// * 'sbcglobal.net'
// * 'freeserve.co.uk'
// * 'orange.net'
// * 'wanadoo.co.uk'
// * 'itelefonica.com.br'
// It's possible these have been shut down. Either way we should probably
// encourage users to adopt more modern email solutions.
var commonEmailDomains = []string{
	/* Default domains included */
	"aol.com", "att.net", "comcast.net", "facebook.com", "gmail.com", "gmx.com", "googlemail.com",
	"google.com", "hotmail.com", "hotmail.co.uk", "mac.com", "me.com", "mail.com", "msn.com",
	"live.com", "verizon.net", "yahoo.com", "yahoo.co.uk",

	/* Other global domains */
	"email.com", "fastmail.fm", "games.com" /* AOL */, "gmx.net", "hush.com", "hushmail.com", "icloud.com",
	"iname.com", "inbox.com", "lavabit.com", "love.com" /* AOL */, "outlook.com", "pobox.com", "protonmail.com",
	"rocketmail.com" /* Yahoo */, "safe-mail.net", "wow.com" /* AOL */, "ygm.com", /* AOL */
	"ymail.com" /* Yahoo */, "zoho.com", "yandex.com",

	/* United States ISP domains */
	"bellsouth.net", "charter.net", "cox.net", "earthlink.net", "juno.com",

	/* British ISP domains */
	"btinternet.com", "virginmedia.com", "blueyonder.co.uk", "live.co.uk",
	"ntlworld.com", "o2.co.uk", "sky.com", "talktalk.co.uk", "tiscali.co.uk",
	"virgin.net", "bt.com",

	/* Domains used in Asia */
	"sina.com", "sina.cn", "qq.com", "naver.com", "hanmail.net", "daum.net", "nate.com", "yahoo.co.jp", "yahoo.co.kr", "yahoo.co.id", "yahoo.co.in", "yahoo.com.sg", "yahoo.com.ph", "163.com", "126.com", "aliyun.com", "foxmail.com",

	/* French ISP domains */
	"hotmail.fr", "live.fr", "laposte.net", "yahoo.fr", "wanadoo.fr", "orange.fr", "gmx.fr", "sfr.fr", "neuf.fr", "free.fr",

	/* German ISP domains */
	"gmx.de", "hotmail.de", "live.de", "online.de", "t-online.de" /* T-Mobile */, "web.de", "yahoo.de",

	/* Italian ISP domains */
	"libero.it", "virgilio.it", "hotmail.it", "aol.it", "tiscali.it", "alice.it", "live.it", "yahoo.it", "email.it", "tin.it", "poste.it", "teletu.it",

	/* Russian ISP domains */
	"mail.ru", "rambler.ru", "yandex.ru", "ya.ru", "list.ru",

	/* Belgian ISP domains */
	"hotmail.be", "live.be", "skynet.be", "voo.be", "tvcablenet.be", "telenet.be",

	/* Argentinian ISP domains */
	"hotmail.com.ar", "live.com.ar", "yahoo.com.ar", "fibertel.com.ar", "speedy.com.ar", "arnet.com.ar",

	/* Domains used in Mexico */
	"yahoo.com.mx", "live.com.mx", "hotmail.es", "hotmail.com.mx", "prodigy.net.mx",

	/* Domains used in Brazil */
	"yahoo.com.br", "hotmail.com.br", "outlook.com.br", "uol.com.br", "bol.com.br", "terra.com.br", "ig.com.br", "r7.com", "zipmail.com.br", "globo.com", "globomail.com", "oi.com.br",
}
