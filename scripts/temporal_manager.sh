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
    queue-dfa)
        Temporal queue-dfa 2>&1 | tee --append /var/log/temporal/queue_dfa.log
        ;;
    ipfs-pin-queue)
        Temporal ipfs-pin-queue 2>&1 | tee --append /var/log/temporal/ipfs_pin_queue.log
        ;;
    ipfs-file-queue)
        Temporal ipfs-file-queue 2>&1 | tee --append /var/log/temporal/ipfs_file_queue.log
        ;;
    ipns-entry-queue)
        Temporal ipns-entry-queue 2>&1 | tee --append /var/log/temporal/ipns_entry_queue.log
        ;;
    email-send-queue)
        Temporal email-send-queue 2>&1 | tee --append /var/log/temporal/email_send_queue.log
        ;;
    ipfs-key-creation-queue)
        Temporal ipfs-key-creation-queue 2>&1 | tee --append /var/log/temporal/ipfs_key_creation_queue.log
        ;;
    migrate)
        Temporal migrate 2>&1 | tee --append /var/log/temporal/database_migrate.log
        ;;
    *)
        echo "[ERROR] Invalid command"
        exit 1
        ;;
esac