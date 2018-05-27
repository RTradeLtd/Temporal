#! /bin/bash

mkdir -p /var/log/temporal
mkdir -p /var/log/ipfs
mkdir -p /ipfs
kdir -p /ipfs/ipfs-cluster
mkdir -p /boot_scripts
bash -x ~/go/src/github.com/RTradeLtd/scripts/ipfs_install.sh
bash -x ~/go/src/github.com/RTradeLtd/scripts/cluster_install.sh
bash -x ~/go/src/github.com/RTradeLtd/scripts/prom_node_exporter_install.sh
cp ~/go/src/github.com/RTradeLtd/scripts/temporal_manager.sh /boot_scripts
cp ~/go/src/github.com/RTradeLtd/scripts/cluster_manager.sh /boot_scripts
cp ~/go/src/github.com/RTradeLtd/scripts/ipfs_manager_script.sh /boot_scripts