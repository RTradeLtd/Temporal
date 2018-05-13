#! /bin/bash


# used to faclitate management of ipfs daemons seperate from ipfs clusters

MODE="daemon"
IPFS_PATH=/srv/dev-disk-by-label-Data/ipfs-data
export IPFS_PATH=/srv/dev-disk-by-label-Data/ipfs-data

case "$MODE" in

    daemon)
        ipfs daemon 2>&1 | tee --append /var/log/ipfs/ipfs_daemon.log
        ;;

    cluster-daemon)
        ipfs-cluster-service daemon | tee --append /var/log/ipfs/ipfs_cluster_daemon.log
        ;;
esac
