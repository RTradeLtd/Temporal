#! /bin/bash

sudo mkdir -p /var/log/temporal
sudo mkdir -p /var/log/ipfs
sudo mkdir -p /ipfs
sudo mkdir -p /ipfs/ipfs-cluster
sudo mkdir -p /boot_scripts
sudo bash -x ~/go/src/github.com/RTradeLtd/Temporal/scripts/ipfs_install.sh
sudo bash -x ~/go/src/github.com/RTradeLtd/Temporal/scripts/cluster_install.sh
sudo bash -x ~/go/src/github.com/RTradeLtd/Temporal/scripts/prom_node_exporter_install.sh
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/scripts/temporal_manager.sh /boot_scripts
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/scripts/cluster_manager.sh /boot_scripts
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/scripts/ipfs_manager_script.sh /boot_scripts