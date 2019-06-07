# Contribution Guidelines

The Contribution guidelines will continually evolve as we gain a better
understanding of our development userbase, and development workflows that need
to be reworked.

In the meantime, all code contributions need to be submitted through
a pull request. Please include a short, detailed description of what your pull
request changes, any breaking changes, and if introducing a new feature, why 
it's of use.

When making changes, please refer to our [versioning standards](/VERSIONING.md)
to see how you can minimize the impact of your change.

All code submitted must include tests that handle proper requests, and improper
requests. Code must also be adequately documented.

## Repository Contents

Refer to the [documentation](https://godoc.org/github.com/RTradeLtd/Temporal)
for an overview of Temporal's package structure.

The project also depends on a [range of sub-repositories](https://github.com/search?q=topic%3Atemporal+org%3ARTradeLtd&type=Repositories).

## Development

### Dependencies

This project leverages go modules, and requires go 1.13 or above.

#### Upgrading

See https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies.
In general:

```
go get -u=patch
```

### Building

Dependencies are vendored by default, so there is no need to install anything if you have the Go toolchain. Temporal is built with `go1.11` and is the only gauranteed version of Golang to work.

To build run the following commands:

```bash
$ go get github.com/RTradeLtd/Temporal
$ cd $GOPATH/src/github.com/RTradeLtd/Temporal
$ make
```

Due to our very large dependency tree which is vendored, running the initial `go get` will take a very long time, anywhere from 15 -> 30 minutes depending on your computer processing capabilities, internet connection, and such. This isn't likely to change anytime soon, so patience is greatly appreciated!

### Running Locally

To run the API locally:

```bash
$ make testenv
$ make api
```

### Testing

Most tests can be run using the following commands:

```bash
$ make testenv
$ make test
```

This requires Docker running and docker-compose installed.

To run the full suite of tests, more advanced setup and configuration is required. If all the prerequisites are set up, you can execute the tests using `make test-all`.

Occassionally the test environment make files may not work on your distribution due to variations in ethernet NIC identifiers. This can be solved by editing `testenv/Makefile` and updating the `INTERFACE=eth0` declaration on line 3.

### Linting

The following command will run some lint checks:

```bash
$ make lint
```

This requires golint and spellcheck to be installed.

## Code Style

### BASH Scripts

All new scripts must pass validation by [shellcheck](https://www.shellcheck.net/)

### Golang

For functions where no return values are kept:

```Golang
if _, err = um.ChangeEthereumAddress( ... ); err != nil {
    ...
}
```
