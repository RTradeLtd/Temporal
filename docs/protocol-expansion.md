# Extending Temporal Protocol Support (WIP)

Supporting a new protocol consists of two steps:
1) Updating the API
2) Updaing the RabbitMQ Consumers

## API Guidelines

We support two APIs, a traditional HTTP API, and a gRPC API which is a work in progress. 

### Updating The HTTP API

Within the API class defned in `api/v2/api.go` there are a few properties which need to be updated. 

If the databse needs to be used, generall you should be able to use existing embedded model managers, however in certain situations, particular when a completely new feature is being rolled out, a new model and associated model manager need to be created. This is done within our [database](https://github.com/RTradeLtd/database) package

If existing queues need to be used, the API already has embedded queue managers within `API::Queues`. If a new queue needs to be created, you'll need to embed a new queue publisher within `API::Queues`.

If you are implementing a completely new feature that doesn't build upon existing API calls, a new route file needs to be declared with the format of `api/v2/route_<route-type>.go`. For example, if adding STORJ functionality `api/v2/route_storj.go` is a good choice of name.

The actual implementation will vary depending on how you need to interact with the protocol being added. For new functions, such as STORJ uploads, a private function part of the API interface needs to be declared. The function name is the type of action your want to perform. An example function declaration for STORJ uploads could be `func (api *API) uploadToStorj(c *gin.Context)`.

In order to make the call available from within the API, the route needs to be included, which is done within the `setupRoutes()` function is `api/v2/api.go`.

Unless there is good reason (such as the Lens search engine) *all* API calls should be protected by our authentication middleware.

## Updating The RabbitMQ Consumers

The RabbitMQ system is the primary backend component responsible for handling tasks sent from Publishers (in this case, the API).
Updating the functionality of Consumers involves declaring a new queue name, an exchange if necessary, the message type using by RabbitMQ to communicate tasks, and the actual functions needed to process tasks received from Publishers.

Queue names are declared in `queue/types.go`, and exchanges are declared within `queue/exchanges.go`. When adding a new queue name, the format is `<Queue-Type>Queue`, for example `STORJUploadQueue`.

Message types are also declared within `queue/types.go` and the name format is simply `<Queue-Type>`

When declaring the functions needed, if no other existing file (such as `queue/ipfs.go` are suitable, create a new file, of `queue/<queue-type>.go` such as `queue/storj_upload.go`