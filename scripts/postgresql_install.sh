#! /bin/bash

echo "[INFO] Updating systems"
sudo apt-get update -y
echo "[INFO] Downloading postgresql"
sudo apt-get install postgresql postgresql-contrib -y
