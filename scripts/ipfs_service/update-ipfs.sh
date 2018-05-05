#!/bin/bash

cd "$(dirname "$0")"

if [ ! $GOPATH ]; then
  echo "I'm missing a GOPATH environment variable. Do you have Go configured correctly?"
  exit
fi

if [ "$1" != "--no-download" ]; then
  source lib/download.sh
fi

echo "Stopping the service and copying the binary over. This will require root."
sudo service ipfs stop 
sudo cp $GOPATH/bin/ipfs /usr/local/bin/ipfs
sudo service ipfs start
