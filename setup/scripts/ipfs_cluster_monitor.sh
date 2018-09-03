#! /bin/bash

case "$1" in 

    peer-count)
        COUNT=$(ipfs-cluster-ctl peers ls | grep "Qm.* |" | grep -v "$(hostname)" | grep -vic "error")
        if [[ "$COUNT" -lt 1 ]]; then
            # if there is no peer, then we print a 1, indicating an error, peers in an error state count as no peer
            echo "1"
            exit 1
        else
            echo "0"
            exit 0
        fi
        ;;
    pin-count)
        COUNT=$(ipfs-cluster-ctl pin ls | wc -l)
        echo "$COUNT"
        ;;
esac