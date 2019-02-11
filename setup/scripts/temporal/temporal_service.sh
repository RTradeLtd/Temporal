#! /bin/bash

# used to launch temporal services
/boot_scripts/temporal_manager.sh krab &
# sleep to give krab enough startup time
sleep 10
/boot_scripts/temporal_manager.sh api &
/boot_scripts/temporal_manager.sh ipfs-pin-queue &
/boot_scripts/temporal_manager.sh email-send-queue &
/boot_scripts/temporal_manager.sh ipns-entry-queue &
/boot_scripts/temporal_manager.sh ipfs-key-creation-queue &
/boot_scripts/temporal_manager.sh ipfs-cluster-queue &
