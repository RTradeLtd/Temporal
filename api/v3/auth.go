package v3

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"gopkg.in/dgrijalva/jwt-go.v3"

	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/sdk/go/temporal"
)

// JWTConfig denotes JWT signing configuration
type JWTConfig struct {
	Key   string
	Realm string
}

// AuthService implements TemporalAuthService
type AuthService struct {
	users  *models.UserManager
	emails *queue.Manager

	jwt JWTConfig
	dev bool

	l *zap.SugaredLogger
}

// VerificationHandler is a traditional HTTP handler for handling account verifications
func (a *AuthService) VerificationHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Register returns the Temporal API status
func (a *AuthService) Register(ctx context.Context, req *auth.RegisterReq) (*auth.User, error) {
	var (
		email = req.GetEmailAddress()
		user  = req.GetCredentials().GetUsername()
		pw    = req.GetCredentials().GetPassword()
		l     = a.l.With("user", user, "email", email)
	)

	// validate email
	if strings.ContainsRune(email, '+') {
		return nil, grpc.Errorf(codes.InvalidArgument, "emails must not contain + signs")
	}

	// create account
	if _, err := a.users.NewUserAccount(user, pw, email); err != nil {
		switch err.Error() {
		case eh.DuplicateEmailError:
			return nil, grpc.Errorf(codes.InvalidArgument, eh.DuplicateEmailError)
		case eh.DuplicateUserNameError:
			return nil, grpc.Errorf(codes.InvalidArgument, eh.DuplicateUserNameError)
		default:
			l.Errorw("unexpected error occured while creating account",
				"error", err)
			return nil, grpc.Errorf(codes.InvalidArgument, eh.UserAccountCreationError)
		}
	}

	// generate a random token to validate email
	u, err := a.users.GenerateEmailVerificationToken(user)
	if err != nil {
		l.Errorw(eh.EmailTokenGenerationError, "error", err)
		return nil, grpc.Errorf(codes.Internal, eh.EmailTokenGenerationError)
	}
	// generate a jwt used to trigger email validation
	token, err := a.signChallengeToken(u.UserName, u.EmailVerificationToken)
	if err != nil {
		l.Errorw("failed to generate email verification jwt", "error", err)
		return nil, grpc.Errorf(codes.Internal, "failed to generate email verification jwt")
	}
	// format the url the user clicks to activate email
	url := fmt.Sprintf("https://api.temporal.cloud/v3/auth/verify?user=%s&challenge%s", u.UserName, token)
	if a.dev {
		url = fmt.Sprintf("https://dev.api.temporal.cloud/v3/auth/verify?user=%s&challenge=%s", u.UserName, token)
	}
	// send email message to queue for processing
	if err = a.emails.PublishMessage(queue.EmailSend{
		Subject: "Temporal Email Verification",
		Content: fmt.Sprintf("please click this %s to activate temporal email functionality",
			fmt.Sprintf("<a href=\"%s\">link</a>", url)),
		ContentType: "text/html",
		UserNames:   []string{u.UserName},
		Emails:      []string{u.EmailAddress},
	}); err != nil {
		l.Errorw(eh.QueuePublishError, "error", err)
		return nil, grpc.Errorf(codes.Internal, "failed to send verification email")
	}
	l.Info("user account registered")

	// return relevant user data
	return &auth.User{
		Id:           uint64(u.ID),
		UserName:     u.UserName,
		EmailAddress: u.EmailAddress,
		Verified:     false,
		Credits:      u.Credits,
		IpfsKeys: func(k []string, v []string) map[string]string {
			m := make(map[string]string)
			for i, key := range k {
				m[key] = v[i]
			}
			return m
		}(u.IPFSKeyIDs, u.IPFSKeyNames),
		IpfsNetworks: u.IPFSNetworkNames,
		Tier:         auth.Tier_FREE,
		ApiAccess:    true,
		AdminAccess:  u.AdminAccess,
	}, nil
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

func (a *AuthService) signChallengeToken(user, challenge string) (string, error) {
	return jwt.
		NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"user":      user,
			"challenge": challenge,
			"expire":    time.Now().Add(time.Hour * 24).UTC().String(),
		}).
		SignedString([]byte(a.jwt.Key))
}
