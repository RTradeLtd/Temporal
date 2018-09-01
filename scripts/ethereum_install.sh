#! /bin/bash

TAG="tags/v1.8.14"

sudo apt install make gcc -y
go get -u -v github.com/ethereum/go-ethereum
cd "$GOPATH/src/github.com/ethereum/go-ethereum" || exit
git checkout "$TAG"
make all