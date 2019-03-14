package v3

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"gopkg.in/dgrijalva/jwt-go.v3"

	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/sdk/go/temporal"
)

// AuthService implements TemporalAuthService
type AuthService struct {
	users *models.UserManager

	l *zap.SugaredLogger
}

// Register returns the Temporal API status
func (a *AuthService) Register(ctx context.Context, req *auth.RegisterReq) (*auth.User, error) {
	return nil, nil
}

// Recover facilitates account recovery
func (a *AuthService) Recover(ctx context.Context, req *auth.RecoverReq) (*auth.User, error) {
	return nil, nil
}

// Login accepts credentials and returns a token for use with further requests.
func (a *AuthService) Login(ctx context.Context, req *auth.Credentials) (*auth.Token, error) {
	return nil, nil
}

// Account returns the account associated with an authenticated request.
func (a *AuthService) Account(ctx context.Context, req *auth.Empty) (*auth.User, error) {
	return nil, nil
}

// Update facilitates modification of the account associated with an
// authenticated request.
func (a *AuthService) Update(ctx context.Context, req *auth.UpdateReq) (*auth.User, error) {
	return nil, nil
}

// Refresh provides a refreshed token associated with an authenticated request.
func (a *AuthService) Refresh(ctx context.Context, req *auth.Empty) (*auth.Token, error) {
	return nil, nil
}

// newAuthInterceptors creates unary and stream interceptors that validate
// requests, for use with gRPC servers
func (a *AuthService) newAuthInterceptors(keyLookup jwt.Keyfunc) (
	unaryInterceptor grpc.UnaryServerInterceptor,
	streamInterceptor grpc.StreamServerInterceptor,
) {
	unaryInterceptor = func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (r interface{}, err error) {
		if ctx, err = a.validate(ctx, keyLookup); err != nil {
			return
		}
		if handler != nil {
			return handler(ctx, req)
		}
		return
	}

	streamInterceptor = func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		var ctx = stream.Context()
		if ctx, err = a.validate(ctx, keyLookup); err != nil {
			return
		}
		if handler != nil {
			return handler(srv, stream)
		}
		return
	}

	return
}

func (a *AuthService) validate(ctx context.Context, keyLookup jwt.Keyfunc) (context.Context, error) {
	// get authorization from context
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok || meta == nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "missing context metadata")
	}
	keys, ok := meta[string(temporal.MetaKeyAuthorization)]
	if !ok || len(keys) == 0 {
		return nil, grpc.Errorf(codes.Unauthenticated, "no key provided")
	}
	var bearerString = keys[0]

	// split out the actual token from the header.
	splitToken := strings.Split(bearerString, "Bearer ")
	if len(splitToken) < 2 {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid key")
	}
	tokenString := splitToken[1]

	// parse takes the token string and a function for looking up the key.
	token, err := jwt.Parse(tokenString, keyLookup)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid key")
	}

	// verify the claims - this checks expiry as well
	var claims jwt.MapClaims
	if claims, ok = token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid key")
	}

	// retrieve ID
	var userID string
	if userID, ok = claims["id"].(string); !ok || userID == "" {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid key")
	}

	// the user should be valid
	var user *models.User
	if user, err = a.users.FindByUserName(userID); err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "invalid user")
	}

	// set user in for retrieval context
	return ctxSetUser(ctx, user), nil
}
