#! /bin/bash

# used to reconnect our rtns publisher



if [[ "$PUB1_ID" == "" ]]; then
    echo "PUB1_ID env var is empty"
    exit 1
fi

if [[ "$PUB2_ID" == "" ]]; then
    echo "PUB2_ID env var is empty"
    exit 1
fi

while true; do

    PUB1_OUT=$(ipfs swarm peers | grep "$PUB1_ID")
    PUB2_OUT=$(ipfs swarm peers | grep "$PUB2_ID")


    # only connect to peer if we arent
    if [[ "$PUB1_OUT" == "" ]]; then
        echo "connecting to node 1"
        /ip4/./tcp/3999/ipfs/"$PUB1_ID"
        /ip4/./tcp/3999/ipfs/"$PUB1_ID"
        /ip4/./tcp/3999/ipfs/"$PUB1_ID"
    fi


    # only connect to peer if we arent
    if [[ "$PUB2_OUT" == "" ]]; then
        echo "connecting to node 2"
        /ip4/./tcp/3999/ipfs/"$PUB2_ID"
        /ip4/./tcp/3999/ipfs/"$PUB2_ID"
        /ip4/./tcp/3999/ipfs/"$PUB2_ID"
    fi

done