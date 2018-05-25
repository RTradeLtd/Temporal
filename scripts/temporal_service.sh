#! /bin/bash

# used to launch temporal services
/boot_scripts/temporal_manager api &
/boot_scripts/temporal_manager queue-dpa &
/boot_scripts/temporal_manager queue-dfa &
/boot_scripts/temporal_manager ipfs-cluster-queue &
