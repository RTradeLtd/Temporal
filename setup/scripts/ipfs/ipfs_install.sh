#! /bin/bash

IPFS_PATH="/ipfs"
PRIVATE_NODE="no"
VERSION="v0.4.19"
export IPFS_PATH="/ipfs"

cd ~ || exit

sudo mkdir /ipfs
sudo chown -R rtrade:rtrade /ipfs
echo "[INFO] Downloading IPFS"
wget "https://dist.ipfs.io/go-ipfs/${VERSION}/go-ipfs_${VERSION}_linux-amd64.tar.gz"
tar zxvf go-ipfs*.gz
rm -- *gz
cd go-ipfs || exit
echo "[INFO] Running ipfs install script"
sudo ./install.sh
echo "[INFO] Initializing ipfs with server and badgerds profile"
ipfs init --profile=server,badgerds >> ~/ipfs_init_log.txt

if [[ "$PRIVATE_NODE" == "yes" ]]; then

    echo "[INFO] Installing ipfs swarm key gen"
    go get -v github.com/Kubuxu/go-ipfs-swarm-key-gen/ipfs-swarm-key-gen
    echo "[INFO] generating swarm key"
    /home/rtrade/go/bin/ipfs-swarm-key-gen > "$IPFS_PATH/swarm.key"
    echo "[INFO] removing bootstrap peers"
    ipfs bootstrap rm --all
    echo "[INFO] setting LIBP2P_FORCE_PNET=1 please set this in your bashrc file to force private networks"
    export LIBP2P_FORCE_PNET=1

fi
