# Architecture

## Data Privacy

We do not perform analytics, or analyzation of content of files uploaded to our system. We will however track file size, file types, and other non-identifying information that is essential to the performance of our service.

## Use Cases

### TODO

## SWARM

We use s
## IPFS

We will operate an IPFS cluster initially consisting of three nodes, two at HQ one off-site. The two nodes at HQ will exist on physically seperated devices to ensure protection against hardware failures. The third node will exist off-site, in case of network or environmental failures data availability will still exist for our customers while we work on bringing back the failed clusters.

Everytime content is uploaded to our system, we add the hashes to our database to create a easily indexable, and backupable listing of file in our system. Periodically we will parse through this database to ensure the files in our clusters are still there. If we notice any discrepancies, missing files will be re-inserted into the cluster

### How It Works

Payment Orchestration is done using smart contracts, see Smart Contracts for more details

Hold up a minute, files, hashes? I thought IPFS was a content addressed data storage system. It is! However we make use of IPFS for two different purposes. One is to ensure that content people want pinned and accessible, but may not be able to pin regularly themselves is still accessibl; In this model you could be the primary storer of your data, and simply want us to act as a secondary host, backup of your content or we could be the primary provider of the content hash, and you the secondary. The possibilities are endless! The second way we make use of IPFS, is for the data storage, and replication of files that people upload to our storage farms. We decided to  go with IPFS due to it's immense capabilities, promising future, and ease of integration into our infrastructure due to a majority of the protocol being built with golang. 

To prevent abuse of the pricing system, even if a file or hash is already pinned on the system, a subsequent pin request from a different user will incur data charges according to how long that file or hash is to be pinned in our system, since that user is also requesting data persistence. In terms of files remaining in our system, the longest pin request is what we follow. 

In order to prevent issues with pinning large files to clusters, and timeouts we first pin to an ipfs node locally, followed by adding that pin to the cluster pinset. Since IPFS Cluster is itself still a working product, we first pin locally to ensure immediate availability, followed by cluster pinning


For now, data uploaded to our system is stored as is. In in interm should you wish for the data to be encrypted and protected against even us accessing your data (which we would absolutely not do to begin wtih) please encrypt the files before upload, or ensure that the content your hash refers to is encrypted. That being said, our IPFS repo exists on encrypted drives, and all backups we do are encryupted.

### Deployment

Currently we aren't using any special deployment techniques, although this is under works.

Current idea is to use:

https://github.com/hsanjuan/ansible-ipfs-cluster

### Node OS Configuration

OS: Ubuntu 16.04.3LTS
CPU: 4 cores
RAM: 8GB
Disks:
    OS = 64GB
    IPFS Repo Size = 1TB

### Security

#### Internal Endpoints

We never exchange unencrypted secrets, even through encrypted communication channels (ssh, https, etc..) we always encrypt the secret prior to transmission with PGP.

Since the endpoint used by cluster peers to communicate is controlled by the `cluster.listen_multiaddress` and defaults to `/ip4/0.0.0.0/tcp/9096` we use port forwarding, and ip filtering to restrict which remote hosts can access this api.

#### HTTP API Endpoints

HTTP API endpoints will be configured with SSL. We are considering adding LibP2P API endpoint support 

These endpints are controlled with `restapi.http_listen_multiaddress` configured to `/ip4/127.0.0.1/tcp/9094` by defaut, however we alter to `/ip4/0.0.0.0/tcp/9094` and use port forwarding and ip filtering to restrict access.

We use SSL and Basic Auth for access to these endpoints.

For more details on this see `https://cluster.ipfs.io/documentation/security/#http-api-endpoints`

#### IPFS and IPFS Proxy Endpoints

IPFS Cluster peers communicate with the IPFS daemon using unauthenticated HTTP auth. In order to ensure that access to the IPFS HTTP API remains only accessible through localhost, we take advantage of the HTTPS IPFS Proxy endpoint provided by the IPFS cluster service, configurable with `ipfshttp.proxy_listen_multiaddress` which we will overide from the default `/ip4/127.0.0.1/tcp/9095` to `/ip4/0.0.0.0/tcp/1337`


### Pricing

Currently we only issue charges in one month intervals. To prevent people from uploading a file to our system and not paying the monthly fee, when a file is added to our system the RTC required to pay will be locked into a smart contract. We will only withdraw from this once a month. Whatever RTC is remaining after the time expires, you are free to withdraw that. If we forget to withdraw our fees, then that's on us not on you!

Should you wish to pay for these services without RTC and use other cryptos, that is also possible but will incur a 5% markup

### Scalability

RabbitMQ will be used to send pin requests to, first it will pin local then pin to cluster. Add information how rabbitmq will be spec'd out

TODO: Need to formulate a better method of pinning files from local node to cluster as per this from the cluster docs:
```
The reason pins (and unpin) requests are queued is because ipfs only performs one pin at a time, while any other requests are hanging in the meantime. All in all, pinning items which are unavailable in the network may create significants bottlenecks (this is a problem that comes from ipfs), as the pin request takes very long to time out. Facing this problem involves restarting the ipfs node.
```

This is an on-going considering for us, and we are always analyzing how to ensure we can scale as much as possible

### Private Networks: TODO

### Scalability Concerns

https://github.com/ipfs/ipfs-cluster/issues/160

### Cluster Boot Strapping Process

First clear pre-existing state data:
`ipfs-cluster-service state clean`

On the peer to be bootstrapped,
`ipfs-cluster-service --bootstrap <existing-peer-multiaddress>`


### Data Persistence And Backups

Default is `~/.ipfs-cluster/ipfs-cluster-data`

The IPFS Cluster Data Directory will be stored on (decide what RAID levels)

Daily ryncs to a non IPFS storage node will be done



## Smart Contracts

We use smart contracts to faciliate the transparent disclosure and verification of files or hashes in our system. We also use smart contracts to facilitate payment for our services in a manner that ensures all parties are happy with the charges, and so that you won't have any unexpected fees/charges.

We use smart contracts to fulfill the following requirements:
    > User Account
    > File Repository
    > Payment Contract


### Smart Contract - User Account

We use keccak256 hashes of files pinned in our system so that users can verify their hashes without us having to expose the content that is stored in our system (all about dat privacy)

This contract will track users in our system, each user will be identified by the following fiels:
    > eth address
    > rtc balance (balance held in contract, allocated to paying for storage)
    > bytes32 array (keccak256(hashes))
    > mapping (bytes32 => bool) keeps track of files they uploaded


The smart contract will perform the following functions:
    > Allow users to deposit any amount of RTC
    > Allow users to withdraw any RTC that isn't locked
    > When user purchases data storage, lock the total amount of RTC required for the duration. No more, no less
    > Expose getters to read various data
    > Allow file repository, to modify data
    > Allow payment contract to modify data

### Smart Contract - File Repository

Each file will be identified by the following:
    > keccak256 hash of the ipfs file hash
    > eth addresses of all uploaders
    > The currently longest file retention period
        > Note if User A uploads for 2 months, and 1.5 months into duration, User B uploads for 1 month, the retention period is updated from 1 month from that date since it will extend past the end of the retention period from user A. 
    > whether or not the file is still pinned in the system

The smart contract will perform the following functions:
    > Allow backend to update the information
    > Allow payment contract to modify data
    > Alter user profile information as appropriate
    > Expose getters to read various data

### Smart Contract - Payment Contract

Payment system architecture design WIP


## Backend Architecture

## Frontend Architecture

When ever someone who hosts files with us, wants to see their decrypted hash names, files uploaded, etc.. We must authenticate them first. To do so, we request a signed messaged from them ,which we must verify. If the signature passes, they can see the files



## Architecture Configuration 

### IPFS Configuration

Initilaize IPFS with the server profile
`ipfs init --profile=server`

For larger repos use Bader datastore:
    > backup ~/.ipfs
    > `ipfs config profile apply badgerds`
    > `ipfs-ds-convert convert`

If using `refs` pinning method disable automatic GC. This id done by default

Increase maximum number of connections using `Swarm.ConnMgr.HighWater`

Reduce `GracePeriod` to 20s

Increase `DataStore.BloomFilterSize` according to repo size (size in bytes)

Set `DataStore.StorageMax` to the appropriate amount of disk you want to dedicate to ipfs repo

### IPFS-Cluster Configuration 

#### Config Tweaks

As the amount of pis  we deal with increase, we will need to increase `cluster.state_sync_interval` and `cluster.ipfs_sync_interval`

Repinning can be very expensive, ad if this becomes and issue (along with taking a cluster to realize a peer is not respodning, thus triggering repins) adjust `monitor.monbasic.check_interval`

To allow peers to have ample time to be restarted and come online, adjust `raft.wait_for_leader_timeout` (30s, 1m)

If network becomes unstable, increase `raft.commit_retries` and `raft.commit_retry_delay` however this can lead to slower failures


for large pinsets increase `raft.snapshot_interval`
for many operations increase `raft.snapshot_threshold`

Set `pin_method`to `refs`
    increase `pin_tracker.maptracker.concurrent_pins` 3->15
            this is how many thigns ipfs should download at the same time

Disable  auto-GC in `go-ipfs`

increase `informer.disk.metric_ttl` as ipfs datastore sizes increase (`5m` or more for larger repos)
    if using `-1` for replication factor, set to a very high number since informers arent used in this case

#### Config Defaults

`cluster.listen_on_multiaddress = /ip4/0.0.0.0/9096`

`cluster.bootstrap = [....]`

`cluster.state_sync_interval = 0m45s`

`cluster.ipfs_sync_interval = 1m30s`

`cluster.monitor_ping_interval = 10s`

we increase to 30s to allow for proper peer init
`cluster.consensus.raft.wait_for_leader_timeout = 30s`

`cluster.consensus.raft.network_timeout = 5s`

`cluster.consensus.raft.commit_retries = 2`

`cluster.consensus.raft.snapshot_interval = 1m0s`

`cluster.api.restapi.listen_multiaddress = /ip4/0.0.0.0/9094`

`cluster.ipfs_connector.ipfshttp.proxy_listen_multiaddress = /ip4/0.0.0.0/1337`

`concurrent_pins = 2` default is 1, maybe set to 2?

leaving `peers` and `bootstrap` to none will cause a single peer mode cluser

#### using `peers`

`peers` should be filled in the following circumstances:
    Working with stable cluster peers, running in known locations
    Working with an automated deployment tools
    You are able to trigger start/stop/restarts for all peers in the cluster with ease

Each entry is a valid peer multiaddress like `/ip4/192.168.1.103/tcp/9096/ipfs/QmQHKLBXfS7hf8o2acj7FGADoJDLat3UazucbHrgxqisim`

Once the peers have booted for the first time, the current peerset will be maintaned by the consensus component and can only be updated by:

    adding new peers, using the bootstrap method
    removing new peers, using the ipfs-cluster-ctl peers rm method


#### using `bootstrap`

`bootstrap` can be used by leaving the `peers` key empty and providing one, or more bootstrap peers


Each entry is a valid peer multiaddress like
`/ip4/192.168.1.103/tcp/9096/ipfs/QmQHKLBXfS7hf8o2acj7FGADoJDLat3UazucbHrgxqisim`

Bootstrap can only be performed with a clean cluster state (ipfs-cluster-service state clean does it)

Bootstrap can only be performed when all the existing cluster-peers are running

You will want to use bootstrap when:

    You are building your cluster manually, starting one single-cluster peer first and boostrapping each node consecutively to it

    You don’t know the IPs or ports your peers will listen to (other than the first)

    Do not manually modify the `peers` (by adding or removing peers) key after the peer has been sucessfully bootstrapped. This will result in startup errors.

`leave_on_shutdown`

The cluster.leave_on_shutdown option allows a peer to remove itself from the peerset when shutting down cleanly:

    The state will be cleaned up automatically when the peer is cleanly shutdown.
    All known peers will be set as bootstrap values and peers will be emptied. Thus, the peer can be started and it will attempt to re-join the cluster it left by bootstrapping to one of the previous peers.
