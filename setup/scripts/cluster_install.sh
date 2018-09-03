#! /bin/bash

# Generate the secret for the cluster

NODE="initial_peer"

# if we are setting up the first node, lets generate a new cluster secret
if [[ "$NODE" == "initial_peer" ]]; then
    CLUSTER_SECRET=$(od  -vN 32 -An -tx1 /dev/urandom | tr -d ' \n')
    export CLUSTER_SECRET
fi

IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster
export IPFS_CLUSTER_PATH=/ipfs/ipfs-cluster

# Install and initialize ipfs-cluster

cd ~ || exit
echo "[INFO] Downloading ipfs cluster"
wget https://dist.ipfs.io/ipfs-cluster-service/v0.5.0/ipfs-cluster-service_v0.5.0_linux-amd64.tar.gz
tar zxvf ipfs-cluster-service*.tar.gz
rm -- *gz
cd ipfs-cluster-service || exit
echo "[INFO] Initializing ipfs cluster configuration"
./ipfs-cluster-service init
sudo cp ipfs-cluster-service /usr/local/bin
cd ~ || exit
echo "[INFO] Downloading ipfs cluster ctl"
wget https://dist.ipfs.io/ipfs-cluster-ctl/v0.5.0/ipfs-cluster-ctl_v0.5.0_linux-amd64.tar.gz
tar zxvf ipfs-cluster-ctl*.tar.gz
rm -- *gz
cd ipfs-cluster-ctl || exit
sudo cp ipfs-cluster-ctl /usr/local/bin
echo "[INFO]  IPFS cluster service installed"