#! /bin/bash

MODE="agent"
DISTRO=$(lsb_release -sc)

cd ~ || exit

wget https://repo.zabbix.com/zabbix/3.4/ubuntu/pool/main/z/zabbix-release/zabbix-release_3.4-1+bionic_all.deb
sudo dpkg -i zabbix-release_3.4-1+bionic_all.deb
sudo apt update -y

if [[ "$DISTRO" == "bionic" && "$MODE" == "server" ]]; then
    sudo add-apt-repository "deb http://archive.ubuntu.com/ubuntu $(lsb_release -sc) main universe restricted multiverse" -y
    sudo add-apt-repository ppa:ondrej/php -y
fi

if [[ "$MODE" == "server" ]]; then
    sudo apt install zabbix-server-pgsql zabbix-frontend-php php-pgsql zabbix-agent  -y
else
    sudo apt install zabbix-agent -y
fi

if [[ "$MODE" == "server" ]]; then
    echo "Installation complete, please see https://www.zabbix.com/download?zabbix=3.4&os_distribution=ubuntu&os_version=bionic&db=PostgreSQL for additional setup instructions"
fi
# sudo -u postgres createuser --pwprompt zabbix
# sudo -u postgres createdb -O zabbix zabbix 

 # zcat /usr/share/doc/zabbix-server-pgsql*/create.sql.gz | sudo -u zabbix psql zabbix 

 # local      DATABASE  USER  METHOD  [OPTIONS]
# host       DATABASE  USER  ADDRESS  METHOD  [OPTIONS]

# psql --host=192.168.1.248 --port=5432 --username=zabbix

