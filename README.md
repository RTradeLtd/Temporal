# RTC-IPFS [![Build Status](https://travis-ci.com/RTradeLtd/RTC-IPFS.svg?token=gDSF5EBqJK8E2W8NbsUS&branch=master)](https://travis-ci.com/RTradeLtd/RTC-IPFS)
Proffesional, enterprise IPFS file hosting paid for in RTC

![](https://i.imgflip.com/29m9ch.jpg)


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

# HOW TO:

Upload file with curl:
curl http://localhost:6767/api/v1/ipfs/add-file -F "file=@/home/solidity/go/src/github.com/RTradeLtd/RTC-IPFS/README.md" -H "Content-Type: multipart/form-data"


# TO DO:

Virus scan files that are uploaded before storage onto ipfs

Write smart contract fo facilitate storage:
    > must contain user accounts
        > each user account must contain a valid balance
        > when uploading a file specify duration, the user accoun must contain enough RTC to pay for the entire duration

Optimize RAFT consensus, and other cluster factors