#! /bin/bash

MODE="server"
DISTRO=$(lsb_release -sc)

cd ~
wget https://repo.zabbix.com/zabbix/3.4/ubuntu/pool/main/z/zabbix-release/zabbix-release_3.4-1+bionic_all.deb
sudo dpkg -i zabbix-release_3.4-1+bionic_all.deb
sudo apt update -y

if [[ "$DISTRO" == "bionic" ]]; then
    sudo add-apt-repository "deb http://archive.ubuntu.com/ubuntu $(lsb_release -sc) main universe restricted multiverse" -y
    sudo add-apt-repository ppa:ondrej/php -y
fi

if [[ "$MODE" == "server" ]]; then
    sudo apt install zabbix-server-pgsql zabbix-frontend-php php-pgsql zabbix-agent  -y
else
    sudo apt install zabbix-agent -y
fi