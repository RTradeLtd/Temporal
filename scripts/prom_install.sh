#! /bin/bash

# This is used to install our prometheus
cd ~
wget https://github.com/prometheus/prometheus/releases/download/v2.2.1/prometheus-2.2.1.linux-amd64.tar.gz
tar zxvf prometheus-*tar.gz
rm *.gz
mkdir /usr/local/bin/prometheus_server
mv ~/prometheus-*/* /usr/local/bin/prometheus_server
cp ~/Temporal/configs/prometheus.yml /usr/local/bin/prometheus_server
cp ~/Temporal/configs/prometheus_server.service /etc/systemd/system
cp ~/Temporal/scripts/prom_server_starth.sh /boot_scripts/prom_server_start.sh
chmod a+x /boot_scripts/prom_server_start.sh
sudo systemctl enable prometheus_server.service
