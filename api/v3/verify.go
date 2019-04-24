package v3

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bobheadxi/res"
	"go.uber.org/zap"
	"gopkg.in/dgrijalva/jwt-go.v3"
)

// VerificationHandler is a traditional HTTP handler for handling account verifications
func (a *AuthService) VerificationHandler(
	l *zap.SugaredLogger,
	users userManager,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			user     = r.URL.Query().Get("user")
			tokenStr = r.URL.Query().Get("token")
			l        = l.With("user", user)
		)

		if user == "" || tokenStr == "" {
			res.R(w, r, res.ErrBadRequest("parameters user, token cannot be empty"))
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unable to validate signing method: %v", token.Header["alg"])
			} else if method != a.jwt.SigningAlgo {
				return nil, errors.New("expect hs512 signing method")
			}
			return []byte(a.jwt.Key), nil
		})
		if err != nil {
			res.R(w, r, res.ErrUnauthorized("invalid token", "error", err))
			return
		}
		if !token.Valid {
			res.R(w, r, res.ErrUnauthorized("invalid token"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			res.R(w, r, res.ErrBadRequest("invalid token claims"))
			return
		}
		if v, ok := claims[claimUser].(string); !ok || v != user {
			res.R(w, r, res.ErrBadRequest("user in token does not match request"))
			return
		}

		// check expiry
		if err := claims.Valid(); err != nil {
			res.R(w, r, res.ErrBadRequest("invalid claims",
				"error", err))
			return
		}

		u, err := a.users.FindByUserName(user)
		if err != nil {
			res.R(w, r, res.ErrNotFound("user not found",
				"user", user))
			return
		}
		challenge, ok := claims[claimChallenge].(string)
		if !ok || challenge != u.EmailVerificationToken {
			res.R(w, r, res.ErrBadRequest("challenge in token is incorrect"))
			return
		}
		if _, err := a.users.ValidateEmailVerificationToken(user, challenge); err != nil {
			l.Errorw("unexpected error when validating user",
				"error", err)
			res.R(w, r, res.ErrInternalServer("unable to validate user", err))
			return
		}

		l.Info("user verified")
		res.R(w, r, res.MsgOK("user verified"))
	}
}
