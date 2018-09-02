#! /bin/bash


case "$1" in

    api)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal api" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    queue-dfa)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal queue-dfa" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-pin-queue)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal ipfs-pin-queue" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-file-queue)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal ipfs-file-queue" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    pin-payment-confirmation-queue)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal pin-payment-confirmation-queue" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    email-send-queue)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal email-send-queue" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipns-entry-queue)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal ipns-entry-queue" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-key-creation-queue)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal ipfs-key-creation-queue" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-cluster-queue)
        PID=$(pgrep -ax Temporal | awk '{print $2" "$3}' | grep "Temporal ipfs-cluster-queue" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;    
    *)
        echo "Bad invocation method"
        exit 1
        ;;

esac