
# Dependencies

> github.com/ipfs/go-ipfs-addr

> https://dist.ipfs.io/#ipfs-cluster-service

> https://dist.ipfs.io/#ipfs-cluster-ctl

> https://github.com/whyrusleeping/ipfs-key

> https://github.com/ipfs/go-ipfs-api

> github.com/gin-gonic/gin

IPFS Cluster Go API https://godoc.org/github.com/ipfs/ipfs-cluster

https://cluster.ipfs.io/developer/api/

https://cluster.ipfs.io/guides/quickstart/
https://github.com/gin-gonic/gin/issues/774

https://cluster.ipfs.io/documentation/internals/

https://www.rabbitmq.com/tutorials/tutorial-two-go.html

# HOW TO:

Upload file with curl:
curl http://localhost:6767/api/v1/ipfs/add-file -F "file=@/home/solidity/go/src/github.com/RTradeLtd/Temporal/README.md" -H "Content-Type: multipart/form-data"


# TO DO:

Virus scan files that are uploaded before storage onto ipfs

Write smart contract fo facilitate storage:
    > must contain user accounts
        > each user account must contain a valid balance
        > when uploading a file specify duration, the user accoun must contain enough RTC to pay for the entire duration

Optimize RAFT consensus, and other cluster factors


# NOTES:

If people give us content hashes, they must be accessible via the internet, otherwise the content will not be able to be pinned