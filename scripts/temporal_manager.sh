#! /bin/bash


# Used to handle system launching of temporal

IPFS_PATH="/ipfs"
IPFS_CLUSTER_PATH="/ipfs/ipfs-cluster"
export IPFS_PATH="$IPFS_PATH"
export IPFS_CLUSTER_PATH="$IPFS_CLUSTER_PATH"
export CONFIG_DAG="/home/rtrade/config.json"

case "$1" in

    api)
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
    payment-register-queue)
        Temporal payment-register-queue 2>&1 | tee --append /var/log/temporal/payment_register_queue.log
        ;;
    payment-received-queue)
        Temporal payment-received-queue 2>&1 | tee --append /var/log/temporal/payment_received_queue.log
        ;;
    migrate)
        Temporal migrate 2>&1 | tee --append /var/log/temporal/database_migrate.log
        ;;
    *)
        echo "[ERROR] Invalid command"
        exit 1
        ;;
esac