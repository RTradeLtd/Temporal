# clamd

Golang Clamd Client

[![Build Status](https://travis-ci.org/baruwa-enterprise/clamd.svg?branch=master)](https://travis-ci.org/baruwa-enterprise/clamd)
[![codecov](https://codecov.io/gh/baruwa-enterprise/clamd/branch/master/graph/badge.svg)](https://codecov.io/gh/baruwa-enterprise/clamd)
[![Go Report Card](https://goreportcard.com/badge/github.com/baruwa-enterprise/clamd)](https://goreportcard.com/report/github.com/baruwa-enterprise/clamd)
[![GoDoc](https://godoc.org/github.com/baruwa-enterprise/clamd?status.svg)](https://godoc.org/github.com/baruwa-enterprise/clamd)
[![MPLv2 License](https://img.shields.io/badge/license-MPLv2-blue.svg?style=flat-square)](https://www.mozilla.org/MPL/2.0/)

## Description

clamd is a Golang library and cmdline tool that implements the
Clamd client protocol used by ClamAV.

## Requirements

* Golang 1.10.x or higher

## Getting started

### Clamd client

The clamd client can be installed as follows

```console
$ go get github.com/baruwa-enterprise/clamd/cmd/clamdscan
```

Or by cloning the repo and then running

```console
$ make build
$ ./bin/clamdscan
```

### Clamd library

To install the library

```console
go get github.com/baruwa-enterprise/clamd
```

You can then import it in your code

```golang
import "github.com/baruwa-enterprise/clamd"
```

### Testing

``make test``

## License

MPL-2.0
