#! /bin/bash


case "$1" in

    krab)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal krab" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    api)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal api" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-pin-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs pin" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    email-send-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue email-send" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipns-entry-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs ipns-entry" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-key-creation-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs key-creation" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-cluster-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs cluster" | grep -iv grep | awk '{print $2}')
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