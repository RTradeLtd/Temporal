#!/bin/bash

cd "$(dirname "$0")"

if [ ! $GOPATH ]; then
  echo "I'm missing a GOPATH environment variable. Do you have Go configured correctly?"
  exit
fi

completion_script="$GOPATH/src/github.com/ipfs/go-ipfs/misc/completion/ipfs-completion.bash"

source lib/download.sh

echo "Installing the service now. This will require root"
sudo cp $completion_script /etc/bash_completion.d/
sudo -E lib/install-service.sh `whoami`
