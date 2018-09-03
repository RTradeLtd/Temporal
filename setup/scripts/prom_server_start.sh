#! /bin/bash

# used to start our prometheus server

CONFIG_FILE="/usr/local/bin/prometheus_server/prometheus.yml"
/usr/local/bin/prometheus_server/prometheus --config.file="$CONFIG_FILE"