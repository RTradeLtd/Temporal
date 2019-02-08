#! /bin/bash

# Commonly used cluster management commands
LOG_PATH="/var/log/ipfs"
export IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster
case "$1" in
    daemon)
        if [[ -f "$LOG_PATH/ipfs_cluster_daemon.log" ]]; then
            echo "[INFO] old log file present, rotating"
            TIMESTAMP=$(date +%s)
            mv "$LOG_PATH/ipfs_cluster_daemon.log" "$LOG_PATH/ipfs_cluster_daemon-$TIMESTAMP.log"
            if [[ $? -ne 0 ]]; then
                echo "[ERROR] failed to rotate log file"
                exit 1
            fi
            echo "[INFO] rotated log file"
        fi
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