#! /bin/bash

# download go ipfs
export IPFS_PATH=/ipfs

echo "[INFO] Creating IPFS Repo Directory"
sudo mkdir -p /ipfs/ipfs-cluster

echo "[INFO] Downloading IPFS"
wget https://dist.ipfs.io/go-ipfs/v0.4.15/go-ipfs_v0.4.15_linux-amd64.tar.gz
tar zxvf go-ipfs*.gz
rm *gz
cd go-ipfs
echo "[INFO] Installing IPFS"
sudo ./install.sh
which ipfs
if [[ "$?" -ne 0 ]]; then
    echo "[ERROR] IPFS Installation Failed"
    exit 1
fi

sudo ipfs init --profile=server,badgerds