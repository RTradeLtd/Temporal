#! /bin/bash

case "$1" in 

    peer-count)
        # this will only work with sha256 based peer ids
        COUNT=$(ipfs-cluster-ctl peers ls | grep -v "$(hostname)" | grep "Qm[a-zA-Z].* |" | awk '{print $1}' | wc -l)
        echo "$COUNT"
        ;;
    queued-pin-count)
        COUNT=$(ipfs-cluster-ctl status --local=true | grep -c PIN_QUEUED)
        echo "$COUNT"
        ;;
    pin-count)
        COUNT=$(ipfs-cluster-ctl pin ls --local=true | wc -l)
        echo "$COUNT"
        ;;
    error-count)
        COUNT=$(ipfs-cluster-ctl status --local=true | grep -c PIN_ERROR)
        echo "$COUNT"
        ;;
    daemon-status)
        # if this is not 0 then there's an error
        # this is used to avoid scenarios where the cluster service is up
        # but the ipfs service is down
        COUNT=$(ipfs-cluster-ctl id | grep -ci error)
        # to maintain consistency with zabbix, emitting a 1 for checks like this
        # indicates an okay, while emitting a 0 indicates an error. Thus we need
        # to perform an if check, such that if the returned error count is 0
        # the node is in a good state, so emit a 1 to indicate everything is a-okay
        if [[ "$COUNT" -eq 0 ]]; then
            echo 1
        else
            echo 0
        fi
        ;;
    *)
        echo "invalid invocation"
        echo "valid commands: peer-count, queued-pin-count, pin-count, error-count, daemon-status"
        exit 1
        ;;

esac