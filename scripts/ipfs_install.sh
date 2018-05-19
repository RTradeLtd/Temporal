#! /bin/bash

cd ~

wget https://dist.ipfs.io/go-ipfs/v0.4.14/go-ipfs_v0.4.14_linux-amd64.tar.gz
tar zxvf go-ipfs*.gz
rm *gz
cd go-ipfs
sudo ./install.sh
ipfs init --profile=server,badgerds >> ~/ipfs_init_log.txt