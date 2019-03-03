# libp2p Daemon

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](http://protocol.ai)
[![](https://img.shields.io/badge/project-libp2p-blue.svg?style=flat-square)](http://libp2p.io/)
[![](https://img.shields.io/badge/freenode-%23libp2p-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%libp2p)
[![GoDoc](https://godoc.org/github.com/libp2p/go-libp2p-daemon?status.svg)](https://godoc.org/github.com/libp2p/go-libp2p-daemon)
[![Coverage Status](https://coveralls.io/repos/github/libp2p/go-libp2p-daemon/badge.svg?branch=master)](https://coveralls.io/github/libp2p/go-libp2p-daemon?branch=master)
[![Build Status](https://travis-ci.org/libp2p/go-libp2p-daemon.svg?branch=master)](https://travis-ci.org/libp2p/go-libp2p-daemon)

> A standalone deployment of a libp2p host, running in its own OS process and installing a set of
  virtual endpoints to enable co-local applications to: communicate with peers, handle protocols,
  interact with the DHT, participate in pubsub, etc. no matter the language they are developed in,
  nor whether a native libp2p implementation exists in that language.

🚧 This project is under active development! 🚧

Check out the [ROADMAP](ROADMAP.md) to see what's coming.

## Install

Note that go1.11 is required.

```sh
$ go get github.com/libp2p/go-libp2p-daemon
$ cd $GOPATH/src/github.com/libp2p/go-libp2p-daemon
$ make
$ p2pd
```

## Usage

Check out the [GoDocs](https://godoc.org/github.com/libp2p/go-libp2p-daemon).

## Language Bindings

Daemon bindings enable applications written in other languages to interact with the libp2p daemon process programmatically, by exposing an idiomatic API that handles the socket dynamics and control protocol.

The following bindings exist so far (if you want yours added, please send a PR):

- Go _(reference implementation)_: see the [p2pclient](p2pclient) package in this repo.
- Python: [py-libp2p-daemon-bindings](https://github.com/mhchia/py-libp2p-daemon-bindings).
- Gerbil: [gerbil-libp2p](https://github.com/vyzo/gerbil-libp2p).

If you wish to implement bindings in a new language, refer to the [spec](specs/README.md) for the daemon control protocol and API.

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/libp2p/go-libp2p-daemon/issues)!

This repository falls under the IPFS [Code of Conduct](https://github.com/ipfs/community/blob/master/code-of-conduct.md).

## License
MIT
