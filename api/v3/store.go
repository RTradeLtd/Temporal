package v3

import (
	"context"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Temporal/api/v3/proto/store"
)

// StoreService implements TemporalStoreService
type StoreService struct {
	l *zap.SugaredLogger
}

// Upload accepts files and directories
func (s *StoreService) Upload(store.TemporalStore_UploadServer) error {
	return nil
}

// Download retrieves an object
func (s *StoreService) Download(*store.DownloadReq, store.TemporalStore_DownloadServer) error {
	return nil
}

// Pin handles new pins and pin extensions
func (s *StoreService) Pin(context.Context, *store.Object) (*store.Empty, error) {
	return nil, nil
}

// Stat retrieves details about an object
func (s *StoreService) Stat(context.Context, *store.Object) (*store.ObjectStats, error) {
	return nil, nil
}

// ListObjects retrieves a list of the authenticated user's objects
func (s *StoreService) ListObjects(context.Context, *store.ListObjectsReq) (*store.ObjectList, error) {
	return nil, nil
}

// Publish publishes a message to the requested topic
func (s *StoreService) Publish(context.Context, *store.Event) (*store.Empty, error) {
	return nil, nil
}

// Subscribe subscribes to messages from the requested topic
func (s *StoreService) Subscribe(*store.Topic, store.TemporalStore_SubscribeServer) error {
	return nil
}

// Keys returns the IPFS keys associated with an authenticated request
func (s *StoreService) Keys(context.Context, *store.Empty) (*store.KeyList, error) {
	return nil, nil
}

// NewKey generates a new IPFS key associated with an authenticated request
func (s *StoreService) NewKey(context.Context, *store.Key) (*store.Empty, error) {
	return nil, nil
}
