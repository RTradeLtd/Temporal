#! /bin/bash

# used to install rabbitmq
ADMIN="admin"
PASS="$1"
DISTRO=$(lsb_release -sc)

if [[ "$PASS" == "" ]]; then
    echo "password not provided as first argument"
    echo "Please set as this is used to create rabbitmq account"
    exit 1
fi

cd ~

if [[ "$DISTRO" == "bionic" ]]; then
    echo "deb https://dl.bintray.com/rabbitmq/debian xenial main" | sudo tee /etc/apt/sources.list.d/bintray.rabbitmq.list
    wget -O- https://www.rabbitmq.com/rabbitmq-release-signing-key.asc | sudo apt-key add -
    wget https://packages.erlang-solutions.com/erlang/esl-erlang/FLAVOUR_1_general/esl-erlang_20.3-1~ubuntu~xenial_amd64.deb
    sudo dpkg -ig "esl-erlang_20.3-1~ubuntu~xenial_amd64.deb"

    if [[ "$?" -ne 0 ]]; then
        sudo apt-get install -f
    fi
fi

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
