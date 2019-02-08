#! /bin/bash

VERSION="5.4.2"
cd ~ || exit
echo "[INFO] Downloading grafana"
wget "https://dl.grafana.com/oss/release/grafana_${VERSION}_amd64.deb"
echo "[INFO] Installing grafana"
sudo dpkg -i grafana_*deb
if [[ $? -eq 1 ]]; then 
    echo "[WARNING] Error occured during grafana installation, attempting to install dependencies"
    sudo apt-get install -f
fi

sudo systemctl daemon-reload
sudo systemctl enable grafana-server
echo "[INFO] Successfully installed grafana"


