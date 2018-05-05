# ipfs-cluster-service

by default cluster configs can be found at `~/.ipfs-cluster`

cluster peers are run with the `ipfs-cluster-service` command.



# Manualy generating a cluster secret


export CLUSTER_SECRET=$(od  -vN 32 -An -tx1 /dev/urandom | tr -d ' \n')


# Manualyl generating a private key and peer id
ipfs-key | base64 -w 0



### 
05:13:56.017  INFO    cluster: IPFS Cluster v0.3.5 listening on: cluster.go:94
05:13:56.017  INFO    cluster:         /ip4/127.0.0.1/tcp/9096/ipfs/QmUcjzWikdGfnD533AsRL4nyyLiESDu7gb9X9HFK5ekqzd cluster.go:97
05:13:56.017  INFO    cluster:         /ip4/10.128.0.2/tcp/9096/ipfs/QmUcjzWikdGfnD533AsRL4nyyLiESDu7gb9X9HFK5ekqzd cluster.go:97
05:13:56.021  INFO    restapi: REST API (libp2p-http): ENABLED restapi.go:396
05:13:56.021  INFO   ipfshttp: IPFS Proxy: /ip4/127.0.0.1/tcp/9095 -> /ip4/127.0.0.1/tcp/5001 ipfshttp.go:182
05:13:56.021  INFO    restapi: REST API (HTTP): /ip4/127.0.0.1/tcp/9094 restapi.go:385
05:13:56.021  INFO    restapi:   - /ip4/127.0.0.1/tcp/9096/ipfs/QmUcjzWikdGfnD533AsRL4nyyLiESDu7gb9X9HFK5ekqzd restapi.go:398

#
01:08:01.215  INFO   ipfshttp: IPFS Proxy: /ip4/127.0.0.1/tcp/9095 -> /ip4/127.0.0.1/tcp/5001 ipfshttp.go:182

# Composite Cluster

With ipfs-cluster, we have a proxy-To-IPFS which allows us to represent a cluster peer as if it were an IPFS daemon itself. The advantage to this is that we can use a cluster peer from a different cluster instead of an IPFS daemon

## Composite Cluster Proxy

IPFS Proxy is an IPFS HTTP API:
    > Some requesta re intercepted and trigger cluster ops
    > non intercepted requests forwarded to teh IPFS daemon attached to the eper

Default listen point:
`/ip4/127.0.0.1/tcp/9095` and is prvodied by the `ipfshttp` cconnector component

>Responses from the proxy mimic the IPFS daemon responses, thus allowing to drop-in this endpoint in places where the IPFS API was used before. For example, you can use the go-ipfs CLI as follows:

    ipfs --api /ip4/127.0.0.1/tcp/9095 pin add <cid>
    ipfs --api /ip4/127.0.0.1/tcp/9095 add myfile.txt
    ipfs --api /ip4/127.0.0.1/tcp/9095 pin rm <cid>
    ipfs --api /ip4/127.0.0.1/tcp/9095 pin ls

    The responses would come from cluster, not from go-ipfs.
    
    "Note that the intercepted endpoints aim to mimic the format and response code from IPFS, but they may lack headers. If you encounter a problem, open an issue so we can address it."


# Security


https://github.com/ipfs/ipfs-cluster/issues/79