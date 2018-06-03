#! /bin/bash

# used to launch temporal services
/boot_scripts/temporal_manager.sh api &
/boot_scripts/temporal_manager.sh queue-dpa &
/boot_scripts/temporal_manager.sh queue-dfa &
/boot_scripts/temporal_manager.sh ipfs-cluster-queue &
/boot_scripts/temporal_manager.sh payment-register-queue &
/boot_scripts/temporal_manager.sh payment-received-queue &
