#! /bin/bash

# used to install a single node ipfs cluster


IPFS_CLUSTER_PATH=/srv/dev-disk-by-label-Data/ipfs-data/cluster-data
export IPFS_CLUSTER_PATH=/srv/dev-disk-by-label-Data/ipfs-data/cluster-data
#CLUSTER_SECRET="....."
#export CLUSTER_SECRET="....."

cd ~
wget https://dist.ipfs.io/ipfs-cluster-service/v0.4.0-rc1/ipfs-cluster-service_v0.4.0-rc1_linux-amd64.tar.gz
tar zxvf ipfs-cluster-service*.tar.gz
rm *gz
cd ipfs-cluster-service
./ipfs-cluster-service init