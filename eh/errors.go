package eh

const (
	// IPFSConnectionError is an error used for ipfs connection failures
	IPFSConnectionError = "failed to connect to ipfs"
	// PrivateNetworkAccessError is used for invalid access to private networks
	PrivateNetworkAccessError = "invalid access to private network"
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
	// CostCalculationError is an error message emitted when we are unable to calculate the cost of something
	CostCalculationError = "failed to calculate cost"
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
	// DepositAddressCheckError is an error message given to a user when searching for a deposit address fails
	DepositAddressCheckError = "failed to get deposit address"
	// UserSearchError is an error message given to a user when a username cant be found
	UserSearchError = "unable to find username"
	// CreditRefundError is an error message used when we are unable to refund a users credits
	CreditRefundError = "failed to refund credits for user"
	// IpnsRecordSearchError is an error message given to users when we can't search for any records
	IpnsRecordSearchError = "failed to search for IPNS records, user likely has published none"
	// UnAuthorizedAdminAccess is an error message used whena user attempts to access an administrative route
	UnAuthorizedAdminAccess = "user is not an administrator"
	// DuplicateEmailError is an error used when a user attempts to register with an already taken email address
	DuplicateEmailError = "email address already taken"
	// DuplicateUserNameError is an error used whe na user attempts to register with an already taken user name
	DuplicateUserNameError = "username is already taken"
	// UnableToSaveUserError is an error that occurs when saving the user account
	UnableToSaveUserError = "saving user account to database failed"
	// EmailVerificationError is an error used when a user fails to validate their email address
	EmailVerificationError = "failed to verify email address"
	// EmailTokenGenerationError is an error messaged used when failing to generate a token
	EmailTokenGenerationError = "failed to generate email verification token"
	// ZoneSearchError is an error message used when failing to search for a zone
	ZoneSearchError = "failed to search for zone"
	// RecordSearchError is an error message used when failing to search for a record
	RecordSearchError = "failed to search for record"
	// IPFSDagGetError is an error message when failing to retrieve a dag from ipfs
	IPFSDagGetError = "failed to get dag from ipfs"
	// InvalidObjectIdentifierError is a generic error to indicate that the object identifier that was provided is invalid
	InvalidObjectIdentifierError = "object identifier is of an invalid format"
	// InvalidObjectTypeError is an error message when a user submits an incorrect type to be indexed
	InvalidObjectTypeError = "object type is invalid, must be ipld"
	// FailedToIndexError is an error message when a lens index request fails
	FailedToIndexError = "an error occurred while trying to index this object"
	// FailedToSearchError is an error message when a lens search request fails
	FailedToSearchError = "an error occurred while submitting your search to lens"
	// NoSearchResultsError is an error message used when no search results were returned
	NoSearchResultsError = "there were no entries matching your search query"
	// ChainRiderAPICallError is an error message used when a call to chainrider api fails
	ChainRiderAPICallError = "failed to call chainrider api"
	// KeyExportError is an error messaged used if a key export request fails
	KeyExportError = "failed to export key"
	// PasswordResetError is an error message used when an error occurins during password reset
	PasswordResetError = "failed to reset password"
	// NoAPITokenError is an error when we can't properly validate the jwt token
	NoAPITokenError = "invalid token provided"
	// CantUploadError is an error when a user tries to upload more data than their monthly limit
	CantUploadError = "uploading would breach monthly data limit, please upload a smaller object"
	// DataUsageUpdateError is an error when a failure occurs while trying to update a users
	// current data usage
	DataUsageUpdateError = "an error occured while updating your account data usage"
	// TierUpgradeError is an error when a failure to upgrade a user tier occurs
	TierUpgradeError = "an error occured upgrading your tier"
	// EncryptionError is an error when a failure to encrypt data occurs
	EncryptionError = "an error occured when trying to encrypt file"
	// DatabaseUpdateError is an error message used when a failure to update the database happesn
	DatabaseUpdateError = "en error occured wile updating the database"
	// PinExtendError is an error message used when someone attempts to extend the pin for content they haven't uploaded
	PinExtendError = "failed to extend pin duration, this likely means you haven't actually uploaded this content before"
	// MaxHoldTimeError is an error message when the current hold time value would breach set pin time limits
	MaxHoldTimeError = "a hold time of this long would result in a longer maximum pin time of 2 years, please reduce your hold time and try again"
)
