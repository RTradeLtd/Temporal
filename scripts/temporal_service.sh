#! /bin/bash

# used to launch temporal services
/boot_scripts/temporal_manager.sh api &
/boot_scripts/temporal_manager.sh queue-dfa &
/boot_scripts/temporal_manager.sh ipfs-pin-queue &
/boot_scripts/temporal_manager.sh ipfs-file-queue &
# /boot_scripts/temporal_manager.sh pin-payment-confirmation-queue &
# /boot_scripts/temporal_manager.sh pin-payment-submission-queue &
/boot_scripts/temporal_manager.sh email-send-queue &
/boot_scripts/temporal_manager.sh ipns-entry-queue &
# /boot_scripts/temporal_manager.sh ipfs-pin-removal-queue &
/boot_scripts/temporal_manager.sh ipfs-key-creation-queue &
/boot_scripts/temporal_manager.sh ipfs-cluster-queue &
