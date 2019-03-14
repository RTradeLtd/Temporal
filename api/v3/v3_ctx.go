package v3

import (
	"context"

	"github.com/RTradeLtd/database/models"
)

type internalCtxKey int

const (
	ctxKeyUser internalCtxKey = iota + 1
)

func ctxSetUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, ctxKeyUser, user)
}

func ctxGetUser(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(ctxKeyUser).(*models.User)
	return user, ok
}
