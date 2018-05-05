#! /bin/bash

# Commonly used cluster management commands

case "$1" in

    # used by a node to add itself to a cluster
    bootstrap)
        echo "enter ip of node"
        read -r ip
        echo "enter  cluster peer id"
        read -r peerID
        ipfs-cluster-service --bootstrap "/ip4/$ip/9096/ipfs/$peerID"
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
        ecccho "etner pid cid"
        read -r pinCID
        ipfs-cluster-ctl recover "$pinCID"
        ;;
esac