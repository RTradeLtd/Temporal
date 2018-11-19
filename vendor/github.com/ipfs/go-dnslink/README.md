# go-dnslink

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](http://ipn.io)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](http://ipfs.io/)

> dnslink resolution in go-ipfs

## Table of Contents

- [Background](#background)
- [Install](#install)
- [Usage](#usage)
  - [As a library](#as-a-library)
  - [As a commandline tool](#as-a-commandline-tool)
- [Contribute](#contribute)
  - [Want to hack on IPFS?](#want-to-hack-on-ipfs)
- [License](#license)

## Background

Package dnslink implements a DNS link resolver. dnslink is a basic
standard for placing traversable links in DNS itself. See dnslink.info

A dnslink is a path link in a DNS TXT record, like this:

```
dnslink=/ipfs/QmR7tiySn6vFHcEjBeZNtYGAFh735PJHfEMdVEycj9jAPy
```

For example:

```
> dig TXT ipfs.io
ipfs.io.  120   IN  TXT  dnslink=/ipfs/QmR7tiySn6vFHcEjBeZNtYGAFh735PJHfEMdVEycj9jAPy
```

This package eases resolving and working with thse DNS links. For example:

```go
import (
  dnslink "github.com/jbenet/go-dnslink"
)

link, err := dnslink.Resolve("ipfs.io")
// link = "/ipfs/QmR7tiySn6vFHcEjBeZNtYGAFh735PJHfEMdVEycj9jAPy"
```

It even supports recursive resolution. Suppose you have three domains with
dnslink records like these:

```
> dig TXT foo.com
foo.com.  120   IN  TXT  dnslink=/ipns/bar.com/f/o/o
> dig TXT bar.com
bar.com.  120   IN  TXT  dnslink=/ipns/long.test.baz.it/b/a/r
> dig TXT long.test.baz.it
long.test.baz.it.  120   IN  TXT  dnslink=/b/a/z
```

Expect these resolutions:

```go
dnslink.ResolveN("long.test.baz.it", 0) // "/ipns/long.test.baz.it"
dnslink.Resolve("long.test.baz.it")     // "/b/a/z"

dnslink.ResolveN("bar.com", 1)          // "/ipns/long.test.baz.it/b/a/r"
dnslink.Resolve("bar.com")              // "/b/a/z/b/a/r"

dnslink.ResolveN("foo.com", 1)          // "/ipns/bar.com/f/o/o/"
dnslink.ResolveN("foo.com", 2)          // "/ipns/long.test.baz.it/b/a/r/f/o/o/"
dnslink.Resolve("foo.com")              // "/b/a/z/b/a/r/f/o/o"
```

## Install

```sh
go get github.com/ipfs/go-dnslink
```

## Usage

### As a library

```go
import (
  log
  fmt

  dnslink "github.com/jbenet/go-dnslink"
)

func main() {
  link, err := dnslink.Resolve("ipfs.io")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(link) // string path
}
```

### As a commandline tool

Check out [the commandline tool](dnslink/), which works like this:

```sh
> dnslink ipfs.io
/ipfs/QmR7tiySn6vFHcEjBeZNtYGAFh735PJHfEMdVEycj9jAPy
```

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/ipfs/go-dnslink/issues)!

This repository falls under the IPFS [Code of Conduct](https://github.com/ipfs/community/blob/master/code-of-conduct.md).

### Want to hack on IPFS?

[![](https://cdn.rawgit.com/jbenet/contribute-ipfs-gif/master/img/contribute.gif)](https://github.com/ipfs/community/blob/master/contributing.md)

## License

[MIT](LICENSE) © Juan Benet-Batiz

