#! /bin/bash

# used to install rabbitmq
ADMIN="admin"
PASS="$1"
DISTRO=$(lsb_release -sc)

if [[ "$DISTRO" != "bionic" ]]; then
    echo "[WARN] non bionic distro detected, this installation may not work without adding specific repositories"
    echo "[WARN] Sleeping for 10 seconds before continuing, hit CTRL+C to exit"
    sleep 10
fi

if [[ "$PASS" == "" ]]; then
    echo "password not provided as first argument"
    echo "Please set as this is used to create rabbitmq account"
    exit 1
fi

cd ~ || exit
sudo apt-get update -y
sudo apt-get install rabbitmq-server
sudo systemctl start rabbitmq-server.service
sudo systemctl enable rabbitmq-server.service

# Create user account information
sudo rabbitmqctl add_user "$ADMIN" "$PASS"
sudo rabbitmqctl set_user_tags "$ADMIN" administrator
sudo rabbitmqctl set_permissions -p / "$ADMIN" ".*" ".*" ".*"

# enable the management console
sudo rabbitmq-plugins enable rabbitmq_management
sudo chown -R rabbitmq:rabbitmq /var/lib/rabbitmq/
