#! /bin/bash


# expects an environment variable IPFS_DATASTORE_PATH
# this variable indicates the directory we use for the datastore

export IPFS_DATASTORE_PATH=/ipfs-data
export IPFS_PATH=/ipfs-node

case "$1" in

    writable-datastore)
        if [[ -w "$IPFS_DATASTORE_PATH" ]]; then
           echo 1
        else
           echo 0
        fi
        ;;
    peer-count)
        COUNT=$(ipfs swarm peers | wc -l)
        echo "$COUNT"
        ;;
    inbound-peers)
        COUNT=$(ipfs swarm peers --direction | grep -c inbound)
        echo "$COUNT"
        ;;
    outbound-peers)
        COUNT=$(ipfs swarm peers --direction | grep -c outbound)
        echo "$COUNT"
        ;;
    daemon-status)
        # neat trick avoiding the `| grep -v | expr`
        # https://unix.stackexchange.com/questions/74185/how-can-i-prevent-grep-from-showing-up-in-ps-results
        COUNT=$(ps aux | grep -c "[i]pfs daemon")
        echo "$COUNT"
        ;;
    *)
        echo "invalid invocation"
        echo "valid comands: writable-datastore, peer-count, inbound-peers, outbound-peers, daemon-status"
        exit 1
        ;;

esac