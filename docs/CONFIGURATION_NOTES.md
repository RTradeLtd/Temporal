## Architecture Configuration 

For people who are interested, the following will detail some of our configurations in a note like fashion.

### IPFS Configuration

Initilaize IPFS with the server profile:
 * `ipfs init --profile=server`

For larger repos use Bader datastore:
* backup ~/.ipfs
* `ipfs config profile apply badgerds`
* `ipfs-ds-convert convert`

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
