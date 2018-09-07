package api

const (
	// IPFSConnectionError is an error used for ipfs connection failures
	IPFSConnectionError = "failed to connect to ipfs"
	// PrivateNetworkAccessError is used for invalid access to private networks
	PrivateNetworkAccessError = "invalid access to private netowrk"
	// APIURLCheckError is an error ussed when failing to retrieve an api url
	APIURLCheckError = "failed to get api url"
	// IPFSCatError is an error used when failing to can an ipfs file
	IPFSCatError = "failed to execute ipfs cat"
	// IPFSObjectStatError is an error used when failure to execute object stat occurs
	IPFSObjectStatError = "failed to execute ipfs object stat"
	// IPFSPubSubPublishError is an error message used whe nfailing to publish pubsub msgs
	IPFSPubSubPublishError = "failed to publish pubsub message"
	// UploadSearchError is a error used when searching for uploads fails
	UploadSearchError = "failed to search for uploads in database"
	// NetworkSearchError is an error used when searching for networks fail
	NetworkSearchError = "faild to search for networks"
	// NetworkCreationError is an error used when creating networks in database fail
	NetworkCreationError = "failed to create network"
	// QueueInitializationError is an error used when failing to connect to the queue
	QueueInitializationError = "failed to initialize queue"
	// QueuePublishError is a message used when failing to publish to queue
	QueuePublishError = "failed to publish message to queue"
	// KeySearchError is an error used when failing to search for a key
	KeySearchError = "failed to search for key"
	// KeyUseError is an error used when attempting to use a key the user down ot own
	KeyUseError = "user does not own key"
	// IPFSPinParseError is an error used when failure to parse ipfs pins occurs
	IPFSPinParseError = "failed to parse ipfs pins"
	// IPFSAddError is an error used when failing to add a file to ipfs
	IPFSAddError = "failed to add file to ipfs"
	// FileOpenError is an error used when failing to open a file
	FileOpenError = "failed to open file"
	// MinioPutError is an error used when storing a file in minio
	MinioPutError = "failed to store object in minio"
	// MinioConnectionError is an error used when connecting to minio
	MinioConnectionError = "failed to connect to minio"
)
