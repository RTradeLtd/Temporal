#! /bin/bash

# This file is used to build the ipfs docker container
# and setup all associated functionality

case "$1" in

    network-create)
        # create a user defined network
        sudo docker network create \
            --driver=bridge \
            --subnet=172.67.67.0/24 \
            --ip-range=172.67.67.0/24 \
            --gateway=172.67.67.1 \
            temporal-net
        ;;

    build-ipfs-container)
        # used to build the ipfs container
        sudo docker build -t ipfs .
        ;;

    *)
        echo "Usage: build.sh [network-create|build-ipfs-container]"
        exit 1
esac