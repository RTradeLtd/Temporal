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
        INIT_DB=true
        if [[ "$TEMPORAL_PRODUCTION" == "yes" ]]; then
            GIN_MODE=release
            export GIN_MODE
        fi
        export INIT_DB
        temporal api
        ;;
    queue-dfa)
        INIT_DB=true
        export INIT_DB  
        temporal queue dfa
        ;;
    ipfs-pin-queue)
        INIT_DB=true
        export INIT_DB
        temporal queue ipfs pin
        ;;
    email-send-queue)
        INIT_DB=true
        export INIT_DB
        temporal queue email-send
        ;;
    ipns-entry-queue)
        INIT_DB=true
        export INIT_DB
        temporal queue ipfs ipns-entry
        ;;
    ipfs-key-creation-queue)
        INIT_DB=true
        export INIT_DB
        temporal queue ipfs key-creation
        ;;
    ipfs-cluster-queue)
        INIT_DB=true
        export INIT_DB
        temporal queue ipfs cluster
        ;;
    krab)
        temporal krab
        ;;
    migrate)
        temporal migrate
        ;;
    *)
        echo "[ERROR] Invalid command"
        exit 1
        ;;
esac
