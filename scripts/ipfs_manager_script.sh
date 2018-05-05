#! /bin/bash


# used to faclitate management of ipfs daemons seperate from ipfs clusters

MODE="daemon"


case "$MODE" in

    daemon)
        sudo ipfs daemon 2>&1 | tee --append ipfs.log
        ;;
esac
