#! /bin/bash


# Used to handle system launching of temporal

IPFS_PATH="/srv/dev-disk-by-label-Data/ipfs-data"
export IPFS_PATH="$IPFS_PATH"
export DB_PASS="password123"
export CERT_PATH="/root/certificates/api.pem"
export KEY_PATH="/root/certificates/api.key"
export ADMIN_USER="admin"
export ADMIN_PASS="minutemaid"
export JWT_KEY="test key"
case "$1" in

    api)
        LISTEN_ADDRESS="192.168.1.252"
        export LISTEN_ADDRESS="$LISTEN_ADDRESS"
        Temporal api 2>&1 | tee --append /var/log/temporal/api.log
        ;;
    queue-dpa)
        Temporal queue-dpa 2>&1 | tee --append /var/log/temporal/queue_dpa.log
        ;;
    queue-dfa)
        Temporal queue-dfa 2>&1 | tee --append /var/log/temporal/queue_dfa.log
        ;;
    ipfs-cluster-queue)
        Temporal ipfs-cluster-queue 2>&1 | tee --append /var/log/temporal/ipfs_cluster_queue.log
        ;;
    migrate)
        Temporal migrate 2>&1 | tee --append /var/log/temporal/database_migrate.log
        ;;

esac