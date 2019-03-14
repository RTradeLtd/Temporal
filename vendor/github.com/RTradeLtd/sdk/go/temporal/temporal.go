package temporal

// MetaKey denotes context metadata keys used by the Temporal gRPC API
type MetaKey string

const (
	// MetaKeyAuthorization is the key for authorization tokens
	MetaKeyAuthorization MetaKey = "authorization"
)
