package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// SecureRequestContext attaches given key as metadata to the provided context
func SecureRequestContext(ctx context.Context, key string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"key", key,
	))
}
