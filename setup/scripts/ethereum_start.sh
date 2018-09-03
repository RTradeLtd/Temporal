#! /bin/bash

RPC_IP="0.0.0.0"
RPC_PORT="8545"

geth --nousb --v5disc --maxpeers 200 --metrics --shh --lightpeers 75 --lightserv 40 --rpc --rpcaddr "$RPC_IP" --rpcport "$RPC_PORT" --rpcapi 'admin,db,eth,debug,miner,net,shh,txpool,personal,web3' --rpccorsdomain "*"