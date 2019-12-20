#! /bin/bash

# used to monitor temporal services, echo'ing a 1 on service up, a 0 on service down

case "$1" in
    
    # temporal related service monitors
    krab)
        COUNT=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep -c "krab")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    api)
        COUNT=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep -c "api")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else
            echo 1
        fi
        ;;
    ipfs-pin-queue)
        COUNT=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep -c "queue ipfs pin")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    email-send-queue)
        COUNT=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep -c "queue email-send")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    ipns-entry-queue)
        COUNT=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep -c "queue ipfs ipns-entry")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    ipfs-key-creation-queue)
        COUNT=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep -c "queue ipfs key-creation")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    ipfs-cluster-queue)
        COUNT=$(pgrep -ax temporal | awk '{print $2" "$3" "$4" "$5" "$6" "$7}' | grep -c "queue ipfs cluster")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;

    # pay related service monitors
    pay-eth-queue)
        COUNT=$(pgrep -ax pay | awk '{print $(NF-2)" "$NF}' | grep -c "queue ethereum")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    pay-dash-queue)
        COUNT=$(pgrep -ax pay | awk '{print $(NF-2)" "$NF}' | grep -c "queue dash")
        if [[ "$COUNT" == "0" ]]; then
            echo 0
        else   
            echo 1
        fi
        ;;
    pay-signer-service)
        # $(NF-1) in awk is the second last column
        COUNT=$(pgrep -ax pay | awk '{print $(NF-1)" "$NF}' | grep -c "grpc server")
        if [[ "$COUNT" == "0" ]]; then
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
