#! /bin/bash

export IPFS_PATH="/ipfs"
IPFS_PATH="/ipfs"
ipfs daemon --offline &
go get -u github.com/Kubuxu/go-ipfs-swarm-key-gen/ipfs-swarm-key-gen
ipfs-swarm-key-gen > "$IPFS_PATH/swarm.key"
cp  "$IPFS_PATH/swarm.key" ~/swarm.key

ipfs bootstrap rm --all