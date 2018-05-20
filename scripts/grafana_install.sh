#! /bin/bash

# this is used to install our grafana server
cd ~
wget https://s3-us-west-2.amazonaws.com/grafana-releases/release/grafana_5.1.3_amd64.deb 
sudo dpkg -i grafana_*deb
if [[ $? -eq 1 ]]; then 
    sudo apt-get install -f
fi
 sudo systemctl daemon-reload
 sudo systemctl enable grafana-server
