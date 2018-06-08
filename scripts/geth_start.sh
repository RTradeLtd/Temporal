#! /bin/bash

NETWORK="rinkeby"
DATADIR="/$NETWORK"
LOG="/var/log/geth.log"
/usr/local/bin/geth "--$NETWORK" --metrics --v5disc --maxpeers 150 --syncmode fast --datadir="$DATADIR" 2>&1 | tee --append "$LOG"
