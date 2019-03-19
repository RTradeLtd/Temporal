# Temporal API V2

[API Reference](https://gateway.temporal.cloud/ipns/docs.api.temporal.cloud)

## Reverse Proxy IPFS API

For ease of use, we expose a reverse proxy that allows users to talk directly to an IPFS node using the [IPFS HTTP API](https://docs.ipfs.io/reference/api/http/).

By utilizing this, you can now point *any* client that follows the `IPFS HTTP API` specification, and leverage the same nodes that power our infrastructure, without having to learn how to specifically use Temporal's V2 API. This means that any DApps you build which may leverage existing IPFS API tooling, can now use our infrastructure.

Currently we only expose a subset of `IPFS HTTP API`commands, and restrict users from accessing commands that involve actions like removing data, listing pinned data, shutting down the nodes, etc... 

For example, using the following javascript code you can connect to ipfs using [js-ipfs-http-client](https://github.com/ipfs/js-ipfs-http-client) without needing to worry about using Temporal's API!

```javascript
var ipfsClient = require('ipfs-http-client')
var temporalClient = require('temporal')

// variables
var exampleCID = 'QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv'
// used for authentication
var temporalUser = process.env.TEMPORAL_USER
var temporalPass = process.env.TEMPORAL_PASS
var jwt = process.env.TEMPORAL_JWT

/*
if you need to get the JWT for your account, you can use the following code.

// init temporal client
var tClient = new temporalClient()

tClient.login(temporalUser, temporalPass)
    .then((response => {
        let token = response.token
        console.log(token)
        return token
    }))
    .catch((err => {
        throw err
    }))
*/


var ipfs = ipfsClient({
    host: 'dev.api.temporal.cloud',
    port: '6768',
    'api-path': '/v2/proxy/api/v0/',
    protocol: 'https',
    headers: {
        authorization: 'Bearer ' + jwt
    }
})

console.log("pinning")
ipfs.pin.add(exampleCID, function (err, response) {
    if (err) {
        console.error(err, err.stack)
        throw  err
    }
    console.log(response)
})
```