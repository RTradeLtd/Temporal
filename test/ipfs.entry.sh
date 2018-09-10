#!/bin/sh

# init docker
sudo service docker start

# wait for dockerd to start
while ! sudo docker stats --no-stream ; do
    echo "Waiting for dockerd to launch..."
    sleep 1
done

# pull images
sudo docker pull ipfs/go-ipfs
sudo docker pull ipfs/ipfs-cluster

# start IPFS
sudo docker run --rm -d \
    -p 8080:8080 \
    -p 5001:4001 \
    -p 4001:4001 \
    ipfs/go-ipfs

# start IPFS-cluster
sudo docker run --rm \
    -p 9094:9094 \
    ipfs/ipfs-cluster
