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
	// MinioBucketCreationError is an error used when creating a minio bucket
	MinioBucketCreationError = "failed to create minio bucket"
	// IPFSMultiHashGenerationError is an error used when calculating an ipfs multihash
	IPFSMultiHashGenerationError = "failed to generate ipfs multihash"
	// IPFSClusterStatusError is a error used when getting the status of ipfs cluster
	IPFSClusterStatusError = "failed to get ipfs cluster status"
	// IPFSClusterConnectionError is an error used when connecting to ipfs cluster
	IPFSClusterConnectionError = "failed to connect to IPFS cluster"
	// IPFSClusterPinRemovalError is an error used when failing to remove a pin from the cluster
	IPFSClusterPinRemovalError = "failed to remove pin from cluster"
	// DNSLinkManagerError is an error used when creating a dns link manager
	DNSLinkManagerError = "failed to create dnslink manager"
	// DNSLinkEntryError is an error used when creating dns link entries
	DNSLinkEntryError = "failed to create dns link entry"
	// PaymentCreationError is an error used when creating payments
	PaymentCreationError = "failed to create payment"
	// PinCostCalculationError is an error message used when calculating pin costs
	PinCostCalculationError = "failed to calculate pin cost"
	// PaymentSearchError is an error used when searching for payment
	PaymentSearchError = "failed to search for payment"
	// DuplicateKeyCreationError is an error used when creating a key of the same name
	DuplicateKeyCreationError = "key name already exists"
	// UserAccountCreationError is an error used when creating a user account
	UserAccountCreationError = "failed to create user account"
	// PasswordChangeError is an error used when changing your password
	PasswordChangeError = "failed to change password"
	// NoKeyError is an error message given to a user when they search for keys, but have none
	NoKeyError = "no keys"
	// FileTooBigError is an error message given to a user when attempting to upload a file larger than our limit
	FileTooBigError = "attempting to upload too big of a file"
	// InvalidPaymentTypeError is an error message given to a user when using an invalid payment method
	InvalidPaymentTypeError = "payment type not supported, must be one of: 'eth' 'rtc' 'btc' 'ltc' 'xmr'"
	// InvalidPaymentBlockchainError is an error message given to a user when they provide an invalid blockchain
	InvalidPaymentBlockchainError = "blockchain must be one of: 'ethereum' 'bitcoin' 'litecoin' 'monero'"
	// CreditCheckError is an error messagen given to a user when searching for their credits fails
	CreditCheckError = "failed to search for user credits"
	// InvalidBalanceError is an error message given to a user when they don't have enough credits to pay
	InvalidBalanceError = "user does not have enough credits to pay for api call"
	// CmcCheckError is an error message given to a user when checking cmc fails
	CmcCheckError = "failed to retrieve value from coinmarketcap"
	// DepositAddressCheckError is an error message given to a user when searchign for a deposit address fails
	DepositAddressCheckError = "failed to get deposit address"
	// UserSearchError is an error message given to a user when a username cant be found
	UserSearchError = "unable to find username"
	// CallCostCalculationError is an error message given to a user when we are unable to calculate the cost of their api call
	CallCostCalculationError = "unable to calculate api call cost"
	// CreditRefundError is an error message used when we are unable to refund a users credits
	CreditRefundError = "failed to refund credits for user"
	// UnAuthorizedAdminAccess is an error message used whena user attempts to access an administrative route
	UnAuthorizedAdminAccess = "user is not an administrator"
)
