# Architecture WIP

## Data Privacy

We do not perform analytics, or analyzation of content of files uploaded to our system. We will however track file size, file types, and other non-identifying information that is essential to the performance of our service.

TODO: add more details

## Use Cases

* Paid For IPFS pinning
* Cloud Storage
* Temporary file hosting
* Off-Site Data Storage
* Data Backup

## SWARM

TODO

## Monitoring

For monitoring of the Temporal service, we utilize prometheus as the collection engine, which will plug into Grafana.
For system monitoring two products are being explored:
    * Grafana
    * Zabbix

## IPFS

We will operate an IPFS cluster initially consisting of two nodes, with immediate expansion to three nodes. Each of these ipfs nodes will exist on the pubilc IPFS swarm, however they will only be configured to pin content that is submitted to us. After launch we will be expanding to include private IPFS networks for us, and for clients should you not wish to store your data on the public swarm. Both public and private swarms will be backed by clusters to ensure data availability, and replication.

### How It Works

Payment Orchestration is done using smart contracts. File hosting will be paid for using the Rally Trade Coin (RTC), or with ether. Everytime you wish to pay for data storage, you must submit the necessary payment to a smart contract, along with inputting the name of the hash you wish to pin, or uploading the file through our web interface. After confirming that we have received the payment, we will pin the file to on of our local IPFS nodes. After the file is successfully pin, we pin the hash cluster wide. The reason for doing is that currently the ipfs cluster service is in development, as has some issues with long pin times timing out. By pinning to the local node first, we ensure a high bandwidth low-latency connection between ifps, and the cluster to avoid any delays that might occur due to pinning a hash whose only providing node is located across the globe, and similar situations.

To prevent abuse of the pricing system, even if a file or hash is already pinned on the system, a subsequent pin request from a different user will incur data charges according to how long that file or hash is to be pinned in our system, since that user is also requesting data persistence. In terms of files remaining in our system, the longest pin request is what we follow. 

Data uploaded to our system is stored as is. For example if you were to upload an unencrypted text file, it would be stored unencrypted. If you were to upload an encrypted text file, it would be stored encrypted. That being said, the actual disk drives themselves on which the IPFS repository exists are encryted.

### Node OS Configuration

OS: Ubuntu 16.04.3LTS
CPU: 8 cores
RAM: 16GB

### Security

We provide no direct (internet) access to any of the IPFS and IPFS cluster endpoints that allow invocation of sensitive commands, or node/cluster operations. The API endpoints for Temporal itself (see the `api` package) are the only ones with exposure to the internet. 

Currently we aren't implementing SSL or basic auth on our IPFS and IFPS Cluster endpoints, however since they don't have direct access to the internet, and all communication occurs on a local network we have not deemd this a concern for now.

The interface between the website frontend, and our backend which receives the actual storage requests will be encrypted to prevent traffic sniffing.

#### Internal Endpoints

We never exchange unencrypted secrets, even through encrypted communication channels (ssh, https, etc..) we always encrypt the secret prior to transmission with PGP.

Since the endpoint used by cluster peers to communicate is controlled by the `cluster.listen_multiaddress` and defaults to `/ip4/0.0.0.0/tcp/9096` we use port forwarding, and ip filtering to restrict which remote hosts can access this api.

#### HTTP API Endpoints


These endpints are controlled with `restapi.http_listen_multiaddress` configured to `/ip4/127.0.0.1/tcp/9094` by defaut so we do not implement any special protection measures for now.

We will implement SSL and Basic Auth for access to these endpoints shortly after launch.

For more details on this see `https://cluster.ipfs.io/documentation/security/#http-api-endpoints`

#### IPFS and IPFS Proxy Endpoints

IPFS Cluster peers communicate with the IPFS daemon using unauthenticated HTTP auth. The HTTPS IPFS Proxy endpoint provided by the IPFS cluster service, which communicates with the local ipfs node, configurable with `ipfshttp.proxy_listen_multiaddress` which defaults to`/ip4/127.0.0.1/tcp/9095`, will not have any special configurations.

### Scalability

RabbitMQ is used to distribute workloads in our backend, and will add more nodes to our IPFS clusters as needed. The main issues with IPFS scalability at this point in time exist with very large amount of pins, and data sizes for both ipfs, and ipfs cluster. 

In order to mitigate this, once our clusters begin showing signs of poor performance, sister clusters will be spun up to distribute the workload for new pins, content, etc.. to new clusters.


Scalability Concerns:
* https://github.com/ipfs/ipfs-cluster/issues/160

### Private IPFS Swarms

Details coming soon

### Data Persistence And Backups

We will conduct daily, and weekly backups of our IPFS Repository, and Cluster nodes to restore in case of hardware failure.

## Smart Contracts

We use smart contracts to fullfil the following roles:
* Payment
* User Registration
* File Repository


## High Level Systems Architecture

* We use a Postgresql database to store user account information, and content hashes off-chain, to allow for easy manipulation, indexing, management, and other critical systems administration operations, for which we may not want to incur the latency of blockchain data access for.

* Gin-Gonic is used to create our API

* IPFS + IPFS Cluster are used as the storage backend

* Gorm is used for interacting with our database

* VMWare ESXi is used  as our virtualization software