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
        temporal queue dfa
        ;;
    ipfs-pin-queue)
        temporal queue ipfs pin
        ;;
    ipfs-file-queue)
        temporal queue ipfs file
        ;;
    pin-payment-confirmation-queue)
        temporal queue payment pin-confirmation
        ;;
    pin-payment-submission-queue)
        temporal queue payment pin-submission
        ;;
    email-send-queue)
        temporal queue email-send
        ;;
    ipns-entry-queue)
        temporal queue ipfs ipns-entry
        ;;
    ipfs-key-creation-queue)
        temporal queue ipfs key-creation
        ;;
    ipfs-cluster-queue)
        temporal queue ipfs cluster
        ;;
    migrate)
        temporal migrate
        ;;
    *)
        echo "[ERROR] Invalid command"
        exit 1
        ;;
esac
