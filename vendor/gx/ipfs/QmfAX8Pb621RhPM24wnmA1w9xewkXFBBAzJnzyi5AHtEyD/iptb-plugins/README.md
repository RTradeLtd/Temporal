# IPTB Plugins

This project contains the IPFS plugins for IPTB. Due to the way IPFS manages dependencies,
these plugins cannot be easily loaded into a generic build of IPTB, and must be use with
the IPTB build in this project.

### Example

```
$ iptb auto -type localipfs -count 5
<output removed>

$ iptb start

$ iptb shell 0
$ echo $IPFS_PATH
/home/iptb/testbed/testbeds/default/0

$ echo 'hey!' | ipfs add -q
QmNqugRcYjwh9pEQUK7MLuxvLjxDNZL1DH8PJJgWtQXxuF

$ exit

$ iptb connect 0 4

$ iptb shell 4
$ ipfs cat QmNqugRcYjwh9pEQUK7MLuxvLjxDNZL1DH8PJJgWtQXxuF
hey!
```

### Usage
```
NAME:
   iptb - iptb is a tool for managing test clusters of libp2p nodes

USAGE:
   iptb [global options] command [command options] [arguments...]

VERSION:
   2.0.0

COMMANDS:
     auto     create default testbed and initialize
     testbed  manage testbeds
     help, h  Shows a list of commands or help for one command
   ATTRIBUTES:
     attr  get, set, list attributes
   CORE:
     init     initialize specified nodes (or all)
     start    start specified nodes (or all)
     stop     stop specified nodes (or all)
     restart  restart specified nodes (or all)
     run      run command on specified nodes (or all)
     connect  connect sets of nodes together (or all)
     shell    starts a shell within the context of node
   METRICS:
     logs    show logs from specified nodes (or all)
     events  stream events from specified nodes (or all)
     metric  get metric from node

GLOBAL OPTIONS:
   --testbed value  Name of testbed to use under IPTB_ROOT (default: "default") [$IPTB_TESTBED]
   --quiet          Suppresses extra output from iptb
   --help, -h       show help
   --version, -v    print the version
```

### Install

```
$ go get -d github.com/ipfs/iptb-plugins
$ cd $GOPATH/src/github.com/ipfs/ipfs-plugins
$ make install
```

### License

MIT
