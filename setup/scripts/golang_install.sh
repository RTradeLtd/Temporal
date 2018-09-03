#! /bin/bash

cd ~ || exit

echo "[INFO] Downloading golang"
wget https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz
echo "[INFO] Unpacking golang"
sudo tar -C /usr/local -xzf go1.10.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
mkdir -p ~/go/src/github.com/RTradeLtd
echo "[INFO] Please update the following env variables"
echo "PATH with /usr/local/go/"
echo "GOPATH with $HOME/go"