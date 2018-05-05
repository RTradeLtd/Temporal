#!/bin/bash

export IPFS_PATH=/var/lib/ipfs
IPFS_USER=ipfs-daemon
IPFS_GROUP=ipfs

current_user=$1

ipfs_running="
There's already an ipfs service installed and running. Are you sure you don't
want to run update-ipfs.sh?
"
if service ipfs status &>/dev/null; then
  printf "$ipfs_running"
  select yn in "Exit" "Continue anyways"; do
    case $yn in
      "Exit" ) exit;;
      "Continue anyways" ) service ipfs stop; break;;
    esac
  done
fi

## Check if the daemon user exists
id -u $IPFS_USER &>/dev/null
if [ $? -ne 0 ]; then
  useradd --system $IPFS_USER --shell /bin/false -G fuse
fi

make_mount_dir() {
  dir=$1;
  mkdir $dir;
  chown $IPFS_USER:$IPFS_USER $dir
  chmod 775 $dir
}

## Note: Copying rather than linking avoids permissions problems
cp $GOPATH/bin/ipfs /usr/local/bin/ipfs

## Create the variously required directories
if [ ! -d $IPFS_PATH ]; then mkdir -p $IPFS_PATH; fi
if [ ! -d /ipfs ]; then make_mount_dir /ipfs; fi
if [ ! -d /ipns ]; then make_mount_dir /ipns; fi

echo "Initializing ipfs... This can take some time to generate the keys"
ipfs init >/dev/null
ipfs config Mounts.FuseAllowOther --bool true

egrep -i "^$IPFS_GROUP:" /etc/group &>/dev/null
if [ $? -ne 0 ]; then
  groupadd $IPFS_GROUP
fi

## Add the user to the group if they aren't root
if [ "$current_user" != "root" ]; then
  usermod --append --groups $IPFS_GROUP $current_user
fi

chown -R $IPFS_USER:$IPFS_GROUP $IPFS_PATH

cp init.d/ipfs /etc/init.d/ipfs
chmod +x /etc/init.d/ipfs

## The part that actually tells the system to load the daemon upon start
update-rc.d ipfs defaults >/dev/null

printf "
                           ***********************
****************************  Daemon Installed!  *******************************
                           ***********************
The ipfs daemon has now been installed. A few last things to know:

Importantly, the IPFS_PATH environment variable with the location of the ipfs
configuration must be loaded into your shell when running ipfs commands. I hate
it when scripts touch my .bashrc so I'm going to leave it to you to add into
your .bashrc the following line:
                       export IPFS_PATH=$IPFS_PATH

$current_user has been added to the group $IPFS_GROUP. This gives you permission
to run ipfs commands and edit the config, but that won't take effect until you
log in again.

The daemon has been installed as an init script. After tweaking the ipfs
configuration found at $IPFS_PATH/config to your liking, on most distros,
you may now start the ipfs daemon by running the following:
                        \`sudo service ipfs start\`

"
