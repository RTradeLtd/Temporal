package v3

import (
	"context"

	"github.com/RTradeLtd/database/models"
)

type internalCtxKey int

const (
	ctxKeyUser internalCtxKey = iota + 1
	ctxKeyClaims
)

func ctxSetUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, ctxKeyUser, user)
}

func ctxSetClaims(ctx context.Context, claims map[string]interface{}) context.Context {
	return context.WithValue(ctx, ctxKeyClaims, claims)
}

func ctxGetUser(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(ctxKeyUser).(*models.User)
	return user, ok
}

func ctxGetClaims(ctx context.Context) (map[string]interface{}, bool) {
	claims, ok := ctx.Value(ctxKeyUser).(map[string]interface{})
	return claims, ok
}
