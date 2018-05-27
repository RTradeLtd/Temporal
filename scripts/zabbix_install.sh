#! /bin/bash

cd ~
wget http://repo.zabbix.com/zabbix/3.4/ubuntu/pool/main/z/zabbix-release/zabbix-release_3.4-1+xenial_all.deb
sudo dpkg -i zabbix-release_3.4-1+xenial_all.deb
sudo apt update -y
sudo apt install zabbix-server-pgsql zabbix-frontend-php php-pgsql zabbix-agent  -y