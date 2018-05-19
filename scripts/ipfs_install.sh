#! /bin/bash

cd ~
MODE="daemon"
IPFS_PATH=/srv/dev-disk-by-label-Data/ipfs-data
export IPFS_PATH=/srv/dev-disk-by-label-Data/ipfs-data
wget https://dist.ipfs.io/go-ipfs/v0.4.15/go-ipfs_v0.4.15_linux-amd64.tar.gz
tar zxvf go-ipfs*.gz
rm *gz
cd go-ipfs
sudo ./install.sh
ipfs init --profile=server,badgerds >> ~/ipfs_init_log.txt