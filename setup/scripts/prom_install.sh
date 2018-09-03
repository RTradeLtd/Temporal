#! /bin/bash

# This is used to install our prometheus
cd ~ || exit
echo "[INFO] Downloading prometheus"
wget https://github.com/prometheus/prometheus/releases/download/v2.2.1/prometheus-2.2.1.linux-amd64.tar.gz
echo "[INFO] Unpacking prometheus"
tar zxvf prometheus-*tar.gz
rm -- *.gz
mkdir /usr/local/bin/prometheus_server
mv ~/prometheus-*/* /usr/local/bin/prometheus_server
cp ~/go/src/github.com/RTradeLtd/Temporal/setup/configs/prometheus.yml /usr/local/bin/prometheus_server
cp ~/go/src/github.com/RTradeLtd/Temporal/setup/configs/prometheus_server.service /etc/systemd/system
cp ~/go/src/github.com/RTradeLtd/Temporal/setup/scripts/prom_server_start.sh /boot_scripts/prom_server_start.sh
chmod a+x /boot_scripts/prom_server_start.sh
echo "[INFO] Prometheus installed, enabling service file"
sudo systemctl enable prometheus_server.service
