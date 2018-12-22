#! /bin/bash

# Used to install, and configure IPFS Cluster

IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster
NODE="initial_peer"
export IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster
VERSION="v0.7.0"

# initial peer is responsible for generating a cluster secret
if [[ "$NODE" == "initial_peer" ]]; then
    CLUSTER_SECRET=$(od  -vN 32 -An -tx1 /dev/urandom | tr -d ' \n')
    export CLUSTER_SECRET
elif [[ "$CLUSTER_SECRET" == "" ]]; then
    echo "[ERROR] Please set CLUSTER_SECRET environment variable"
    exit 1
fi

# download and install ipfs cluster service
cd ~ || exit
echo "[INFO] Downloading ipfs cluster"
wget https://dist.ipfs.io/ipfs-cluster-service/v0.7.0/ipfs-cluster-service_v0.7.0_linux-amd64.tar.gz
tar zxvf ipfs-cluster-service*.tar.gz
rm -- *gz
cd ipfs-cluster-service || exit
echo "[INFO] Initializing ipfs cluster configuration"
./ipfs-cluster-service init
sudo cp ipfs-cluster-service /usr/local/bin

# download and install ipfs cluster utility tool
cd ~ || exit
echo "[INFO] Downloading ipfs cluster ctl"
wget https://dist.ipfs.io/ipfs-cluster-ctl/v0.7.0/ipfs-cluster-ctl_v0.7.0_linux-amd64.tar.gz
tar zxvf ipfs-cluster-ctl*.tar.gz
rm -- *gz
cd ipfs-cluster-ctl || exit
sudo cp ipfs-cluster-ctl /usr/local/bin
echo "[INFO]  IPFS cluster service installed"