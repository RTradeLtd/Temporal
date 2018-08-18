#! /bin/bash

# Commonly used cluster management commands
export IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster
case "$1" in
    daemon)
        ipfs-cluster-service daemon 2>&1 | tee --append /var/log/ipfs/ipfs_cluster_daemon.log
        ;;
    # used by a node to add itself to a cluster
    bootstrap)
        echo "enter existing peer multiaddr"
        read -r multiADDR
        ipfs-cluster-service --bootstrap "$multiADDR"
        ;;
    # used to list peers in a cluster
    peer-list)
        ipfs-cluster-ctl peer ls
        ;;
    # used by a node to remove a peer
    remove-peer)
        echo "enter peer id"
        read -r peerID
        ipfs-cluster-ctl peers rm "$peerID"
        ;;
    cluster-status)
        ipfs-cluster-ctl status
        ;;
    pin-details)
        echo "enter pin cid"
        read -r pinCID
        ipfs-cluster-ctl pin ls "$pinCID"
        ;;
    pin-status)
        echo "enter pin cid"
        read -r pinCID
        ipfs-cluster-ctl status "$pinCID"
        ;;
    # used to recover and repin  hashes marked as ERROR in cluster ctl status
    pin-recover)
        echo "etner pid cid"
        read -r pinCID
        ipfs-cluster-ctl recover "$pinCID"
        ;;
    state-clean)
        ipfs-cluster-service state clean
        ;;
esac