#! /bin/bash


# Used to handle system launching of temporal

IPFS_PATH="/ipfs"
IPFS_CLUSTER_PATH="/ipfs/ipfs-cluster"
export IPFS_PATH="$IPFS_PATH"
export IPFS_CLUSTER_PATH="$IPFS_CLUSTER_PATH"
export CONFIG_DAG="/home/rtrade/config.json"

case "$1" in

    api)
        Temporal api
        ;;
    queue-dfa)
        Temporal queue-dfa
        ;;
    ipfs-pin-queue)
        Temporal ipfs-pin-queue
        ;;
    ipfs-file-queue)
        Temporal ipfs-file-queue
        ;;
    pin-payment-confirmation-queue)
        Temporal pin-payment-confirmation-queue
        ;;
    pin-payment-submission-queue)
        Temporal pin-payment-submission-queue
        ;;
    email-send-queue)
        Temporal email-send-queue
        ;;
    ipns-entry-queue)
        Temporal ipns-entry-queue
        ;;
    ipfs-key-creation-queue)
        Temporal ipfs-key-creation-queue
        ;;
    ipfs-cluster-queue)
        Temporal ipfs-cluster-queue
        ;;
    migrate)
        Temporal migrate
        ;;
    *)
        echo "[ERROR] Invalid command"
        exit 1
        ;;
esac