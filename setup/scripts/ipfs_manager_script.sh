#! /bin/bash


# used to faclitate management of ipfs daemons seperate from ipfs clusters

MODE="daemon"
export IPFS_PATH=/ipfs

case "$MODE" in

    daemon)
        ipfs daemon --enable-pubsub-experiment 2>&1 | tee --append /var/log/ipfs/ipfs_daemon.log
        ;;

    cluster-daemon)
        ipfs-cluster-service daemon 2>&1 | tee --append /var/log/ipfs/ipfs_cluster_daemon.log
        ;;
esac
