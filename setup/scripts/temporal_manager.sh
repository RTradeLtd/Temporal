#! /bin/bash


# Used to handle system launching of temporal

IPFS_PATH="/ipfs"
IPFS_CLUSTER_PATH="/ipfs/ipfs-cluster"
CONFIG_DAG="/home/rtrade/config.json"
export IPFS_PATH
export IPFS_CLUSTER_PATH
export CONFIG_DAG

case "$1" in

    api)
        temporal api
        ;;
    queue-dfa)
        temporal queue-dfa
        ;;
    ipfs-pin-queue)
        temporal ipfs-pin-queue
        ;;
    ipfs-file-queue)
        temporal ipfs-file-queue
        ;;
    pin-payment-confirmation-queue)
        temporal pin-payment-confirmation-queue
        ;;
    pin-payment-submission-queue)
        temporal pin-payment-submission-queue
        ;;
    email-send-queue)
        temporal email-send-queue
        ;;
    ipns-entry-queue)
        temporal ipns-entry-queue
        ;;
    ipfs-key-creation-queue)
        temporal ipfs-key-creation-queue
        ;;
    ipfs-cluster-queue)
        temporal ipfs-cluster-queue
        ;;
    migrate)
        temporal migrate
        ;;
    *)
        echo "[ERROR] Invalid command"
        exit 1
        ;;
esac