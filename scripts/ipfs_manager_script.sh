#! /bin/bash


# used to faclitate management of ipfs daemons seperate from ipfs clusters

MODE="daemon"


case "$MODE" in

    daemon)
        sudo ipfs daemon 2>&1 | tee --append /var/log/ipfs/ipfs_daemon.log
        ;;

    cluster-daemon)
        ipfs-cluster-service daemon | tee --append /var/log/ipfs/ipfs_cluster_daemon.log
        ;;
esac
