#! /bin/bash

sudo mkdir -p /var/log/temporal
sudo mkdir -p /var/log/ipfs
sudo mkdir -p /ipfs
sudo mkdir -p /ipfs/ipfs-cluster
sudo mkdir -p /boot_scripts
sudo bash -x ~/go/src/github.com/RTradeLtd/Temporal/scripts/ipfs_install.sh
sudo bash -x ~/go/src/github.com/RTradeLtd/Temporal/scripts/cluster_install.sh
# this is currently broken sudo bash -x ~/go/src/github.com/RTradeLtd/Temporal/scripts/prom_node_exporter_install.sh
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/scripts/temporal_manager.sh /boot_scripts
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/scripts/cluster_manager.sh /boot_scripts
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/scripts/ipfs_manager_script.sh /boot_scripts
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/configs/ipfs.service /etc/systemd/system
sudo cp ~/go/src/github.com/RTradeLtd/Temporal/configs/ipfs_cluster.service /etc/systemd/system
sudo chmod a+x /boot_scripts/ipfs_manager_script.sh
sudo chmod a+x /boot_scripts/cluster_manager.sh
sudo systemctl enable ipfs.service
sudo systemctl enable ipfs_cluster.service