#! /bin/bash

# This is used to install our prometheus
cd ~
wget https://github.com/prometheus/prometheus/releases/download/v2.2.1/prometheus-2.2.1.linux-amd64.tar.gz
tar zxvf prometheus-*tar.gz
rm *.gz
mkdir /usr/local/bin/prometheus_server
mv ~/prometheus-*/* /usr/local/bin/prometheus_server
cp ~/Temporal/configs/prometheus.yml /usr/local/bin/prometheus_server