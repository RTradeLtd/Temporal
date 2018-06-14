#! /bin/bash

cd ~
MODE="daemon"
IPFS_PATH=/ipfs
export IPFS_PATH=/ipfs
wget https://dist.ipfs.io/go-ipfs/v0.4.15/go-ipfs_v0.4.15_linux-amd64.tar.gz
tar zxvf go-ipfs*.gz
rm *gz
cd go-ipfs
bash ./install.sh
ipfs init --profile=server,badgerds >> ~/ipfs_init_log.txt