package dialer

import (
	"context"

	"github.com/RTradeLtd/grpc/middleware"

	"google.golang.org/grpc/metadata"
)

// SecureRequestContext attaches given key as metadata to the provided context
func SecureRequestContext(ctx context.Context, key string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		middleware.AuthorizationKey, key,
	))
}
