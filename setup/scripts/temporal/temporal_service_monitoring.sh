#! /bin/bash

# used to monitor temporal services, echo'ing a 1 on service up, a 0 on service down

case "$1" in
    
    # temporal related service monitors
    krab)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal krab" | grep -iv grep | awk '{print $2}')
        if [[ "$PID" == "krab" ]]; then
            echo 1
        else   
            echo 0
        fi
        ;;
    api)
        # first check production
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal api" | grep -iv grep | awk '{print $(NF-1)" "$NF}')
        if [[ "$PID" == "temporal api" ]]; then
            echo 1
        else
            # if production check fails, check dev, otherwise echo 0   
            PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal -dev api" | grep -iv grep | awk '{print $(NF-2)" "$(NF-1)" "$NF}')
            if [[ "$PID" == "temporal -dev api" ]]; then
                echo 1
            else
                echo 0
            fi
        fi
        ;;
    ipfs-pin-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs pin" | grep -iv grep | awk '{print $(NF-2)" "$(NF-1)" "$NF}')
        if [[ "$PID" == "queue ipfs pin" ]]; then
            echo 1
        else
            echo 0
        fi
        ;;
    email-send-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue email-send" | grep -iv grep | awk '{print $(NF-1)" "$NF}')
        if [[ "$PID" == "queue email-send" ]]; then
            echo 1
        else
            echo 0
        fi
        ;;
    ipns-entry-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs ipns-entry" | grep -iv grep | awk '{print $(NF-2)" "$(NF-1)" "$NF}')
        if [[ "$PID" == "queue ipfs ipns-entry" ]]; then
            echo 1
        else
            echo 0
        fi
        ;;
    ipfs-key-creation-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs key-creation" | grep -iv grep | awk '{print $(NF-2)" "$(NF-1)" "$NF}')
        if [[ "$PID" == "queue ipfs key-creation" ]]; then
            echo 1
        else
            echo 0
        fi
        ;;
    ipfs-cluster-queue)
        PID=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep "temporal queue ipfs cluster" | grep -iv grep | awk '{print $(NF-2)" "$(NF-1)" "$NF}')
        if [[ "$PID" == "queue ipfs cluster" ]]; then
            echo 1
        else
            echo 0
        fi
        ;;

    # pay related service monitors
    pay-eth-queue)
        PID=$(pgrep -ax pay | awk '{print $(NF-2)" "$NF}' | grep "queue ethereum" | grep -iv grep | awk '{print $(NF-1)" "$NF}')
        if [[ "$PID" == "queue ethereum" ]]; then
            echo 1
        else
            echo 0
        fi
        ;;
    pay-dash-queue)
        PID=$(pgrep -ax pay | awk '{print $(NF-2)" "$NF}' | grep "queue dash" | grep -iv grep | awk '{print $(NF-1)" "$NF}')
        if [[ "$PID" == "queue dash" ]]; then
            echo 1
        else
            echo 0
        fi
        ;;
    pay-signer-service)
        # $(NF-1) in awk is the second last column
        PID=$(pgrep -ax pay | awk '{print $(NF-1)" "$NF}' | grep "grpc server" | grep -iv grep | awk '{print $(NF-1)" "$NF}')
        if [[ "$PID" == "grpc server" ]]; then
            echo 1
        else
            echo 0
        fi   
        ;;     
    *)
        echo "Bad invocation method"
        exit 1
        ;;

esac