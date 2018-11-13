package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// AuthorizationKey is the key used to store authorization token data
const AuthorizationKey = "authorization"

// NewServerInterceptors creates unary and stream interceptors that validate
// requests, for use with gRPC servers, using given key
func NewServerInterceptors(key string) (
	unaryInterceptor grpc.UnaryServerInterceptor,
	streamInterceptor grpc.StreamServerInterceptor,
) {
	unaryInterceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, grpc.Errorf(codes.Unauthenticated, "missing context metadata")
		}
		if err := validate(meta, key); err != nil {
			return nil, err
		}
		if handler == nil {
			return nil, nil
		}
		return handler(ctx, req)
	}

	streamInterceptor = func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		meta, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return grpc.Errorf(codes.Unauthenticated, "missing context metadata")
		}
		if err := validate(meta, key); err != nil {
			return err
		}
		if handler == nil {
			return nil
		}
		return handler(srv, stream)
	}

	return
}

func validate(meta metadata.MD, key string) error {
	keys, ok := meta[AuthorizationKey]
	if !ok || len(meta[AuthorizationKey]) == 0 {
		return grpc.Errorf(codes.Unauthenticated, "no key provided")
	}
	if keys[0] != key {
		return grpc.Errorf(codes.Unauthenticated, "invalid key")
	}
	return nil
}
