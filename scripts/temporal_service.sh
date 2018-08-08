#! /bin/bash

# used to launch temporal services
/boot_scripts/temporal_manager.sh api &
/boot_scripts/temporal_manager.sh queue-dfa &
/boot_scripts/temporal_manager.sh ipfs-cluster-queue &
/boot_scripts/temporal_manager.sh ipfs-pin-queue &
/boot_scripts/temporal_manager.sh ipfs-file-queue &
