#!/bin/sh

IPFSVERSION="v0.4.17"
CLUSTERVERSION="v0.5.0"

IPFSIMAGE=ipfs/go-ipfs:$IPFSVERSION
CLUSTERIMAGE=ipfs/ipfs-cluster:$CLUSTERVERSION

# init docker
sudo service docker start

# wait for dockerd to start
while ! sudo docker stats --no-stream ; do
    echo "Waiting for dockerd to launch..."
    sleep 1
done

# pull images
sudo docker pull $IPFSIMAGE
sudo docker pull $CLUSTERIMAGE

# create container network
sudo docker network create container_internal

# start IPFS
sudo docker create \
    --network container_internal \
    --name node \
    -p 8080:8080 \
    -p 5001:5001 \
    -p 4001:4001 \
    $IPFSIMAGE
sudo docker start node

# start IPFS-cluster
sudo docker create \
    --network container_internal \
    --name cluster \
    -p 9094:9094 \
    -p 9095:9095 \
    -v /ipfs-cluster.service.json:/data/ipfs-cluster/service.json \
    $CLUSTERIMAGE "$@"

sudo docker start --attach cluster
