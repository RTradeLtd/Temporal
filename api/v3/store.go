package v3

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/RTradeLtd/Temporal/api/v3/proto/store"
)

// StoreService implements TemporalStoreService
type StoreService struct {
	l *zap.SugaredLogger
}

// Upload accepts files and directories
func (s *StoreService) Upload(store.TemporalStore_UploadServer) error {
	return grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// UploadBlob accepts a single blob (up to 5mb)
func (s *StoreService) UploadBlob(context.Context, *store.UploadReq) (*store.Object, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// Download retrieves an object
func (s *StoreService) Download(*store.DownloadReq, store.TemporalStore_DownloadServer) error {
	return grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// DownloadBlob returns a single blob (up to 5mb)
func (s *StoreService) DownloadBlob(context.Context, *store.DownloadReq) (*store.Blob, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// Pin handles new pins and pin extensions
func (s *StoreService) Pin(context.Context, *store.Object) (*store.Empty, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// Stat retrieves details about an object
func (s *StoreService) Stat(context.Context, *store.Object) (*store.ObjectStats, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// ListObjects retrieves a list of the authenticated user's objects
func (s *StoreService) ListObjects(context.Context, *store.ListObjectsReq) (*store.ObjectList, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// Publish publishes a message to the requested topic
func (s *StoreService) Publish(context.Context, *store.Event) (*store.Empty, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// Subscribe subscribes to messages from the requested topic
func (s *StoreService) Subscribe(*store.Topic, store.TemporalStore_SubscribeServer) error {
	return grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// Keys returns the IPFS keys associated with an authenticated request
func (s *StoreService) Keys(context.Context, *store.Empty) (*store.KeyList, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// NewKey generates a new IPFS key associated with an authenticated request
func (s *StoreService) NewKey(context.Context, *store.Key) (*store.Empty, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}
