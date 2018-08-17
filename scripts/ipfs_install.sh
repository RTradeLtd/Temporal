#! /bin/bash

cd ~ || exit

IPFS_PATH=/ipfs
export IPFS_PATH=/ipfs
sudo mkdir /ipfs
sudo chown -R rtrade:rtrade /ipfs
wget https://dist.ipfs.io/go-ipfs/v0.4.17/go-ipfs_v0.4.17_linux-amd64.tar.gz
tar zxvf go-ipfs*.gz
rm *gz
cd go-ipfs
sudo ./install.sh
ipfs init --profile=server,badgerds >> ~/ipfs_init_log.txt