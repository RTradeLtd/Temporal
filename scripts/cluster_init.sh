#! /bin/bash

# Currently generates single peer cluster

# Generate the secret for the cluster
CLUSTER_SECRET=$(od  -vN 32 -An -tx1 /dev/urandom | tr -d ' \n')

export "$CLUSTER_SECRET"
echo -e "Cluster Secret: $CLUSTER_SECRET\n"

# initialize the configuration
ipfs-cluster-service init

# start the daemon
ipfs-cluster-service daemon