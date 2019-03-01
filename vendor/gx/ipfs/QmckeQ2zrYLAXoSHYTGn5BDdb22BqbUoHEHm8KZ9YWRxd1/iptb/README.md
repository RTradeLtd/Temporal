# IPTB

`iptb` is a program used to create and manage a cluster of sandboxed nodes
locally on your computer. Spin up 1000s of nodes! Using `iptb` makes testing
libp2p networks easy!

For working with IPFS please see [ipfs/iptb-plugins](https://github.com/ipfs/iptb-plugins).

### Example (ipfs)

```
$ iptb auto -type <plugin> -count 5
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

_Note: For MacOS golang v1.11 is needed to support plugin loading
(see [golang/go#24653](https://github.com/golang/go/issues/24653) for more information)_

```
$ go get github.com/ipfs/iptb
```

### Plugins

Plugins are now used to implement support for managining nodes. Plugins are
stored under `$IPTB_ROOT/plugins` (see [configuration](#configuration))

Plugins for the IPFS project can be found in [ipfs/iptb-plugins](https://github.com/ipfs/iptb-plugins).

### Configuration

By default, `iptb` uses `$HOME/testbed` to store created nodes. This path is configurable via the environment variables `IPTB_ROOT`.

### License

MIT
