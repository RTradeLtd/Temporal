IPFS Linux Init Daemon
======================
###
#
# Copied from https://github.com/RTradeLtd/ipfs-linux-service
# 
###

Note that for multi-user systems this is entirely useless. For a single user
system however, it's helpful because IPFS will
automatically be started at boot and always be at the ready without resorting
to running it in a screen process.

Install
-------
To install ipfs as a daemon, simply run:
```
./install-ipfs.sh
```
I also included a script that enables you to easily update ipfs. That can be run
with:
```
./update-ipfs.sh
```
Note that these don't require being ran as root, but will instead later ask 
for root permissions.

You'll be able to later start and stop the daemon with the command:
```
sudo service ipfs start/stop/restart
```
