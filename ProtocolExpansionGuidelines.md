# Extending Temporal Protocol Support (WIP)

Supporting a new protocol consists of two steps:
1) Updating the API
2) Updaing the RabbitMQ Consumers

## Updating The API

## Updating The RabbitMQ Consumers

The RabbitMQ system is the primary backend component responsible for handling tasks sent from Publishers (in this case, the API).
Updating the functionality of Consumers involves declaring a new queue name, an exchange if necessary, the message type using by RabbitMQ to communicate tasks, and the actual functions needed to process tasks received from Publishers.

Queue names are declared in `queue/types.go`, and exchanges are declared within `queue/exchanges.go`. When adding a new queue name, the format is `<Queue-Type>Queue`, for example `STORJUploadQueue`.

Message types are also declared within `queue/types.go` and the name format is simply `<Queue-Type>`

When declaring the functions needed, if no other existing file (such as `queue/ipfs.go` are suitable, create a new file, of `queue/<queue-type>.go` such as `queue/storj_upload.go`