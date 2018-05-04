# ipfs-cluster-service

by default cluster configs can be found at `~/.ipfs-cluster`

cluster peers are run with the `ipfs-cluster-service` command.



# Manualy generating a cluster secret


export CLUSTER_SECRET=$(od  -vN 32 -An -tx1 /dev/urandom | tr -d ' \n')


# Manualyl generating a private key and peer id
ipfs-key | base64 -w 0