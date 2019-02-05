#! /bin/bash


# used to faclitate management of ipfs daemons seperate from ipfs clusters

LOG_PATH="/var/log/ipfs"
MODE="daemon"
export IPFS_PATH=/ipfs

case "$MODE" in

    daemon)
        if [[ -f "$LOG_PATH/ipfs_daemon.log" ]]; then
            echo "[INFO] old log file present, rotating"
            TIMESTAMP=$(date +%s)
            mv "$LOG_PATH/ipfs_daemon.log" "$LOG_PATH/ipfs_daemon-$TIMESTAMP.log"
            if [[ $? -ne 0 ]]; then
                echo "[ERROR] failed to rotate log file"
                exit 1
            fi
            echo "[INFO] rotated log file"
        fi
        ipfs daemon --enable-namesys-pubsub 2>&1 | tee --append /var/log/ipfs/ipfs_daemon.log
        ;;

esac
