<h1 align="center">Temporal ☄️</h1>

<p align="center">
  <a href="#about-temporal"><strong>About Temporal</strong></a> · 
  <a href="#web-interfaces"><strong>Web Interfaces</strong></a> · 
  <a href="#usage-and-features"><strong>Usage and Features</strong></a> · 
  <a href="#development"><strong>Development</strong></a> · 
  <a href="#license"><strong>License</strong></a> · 
  <a href="#thanks"><strong>Thanks</strong></a> ·

</p>

<p align="center">
  <a href="https://godoc.org/github.com/RTradeLtd/Temporal">
    <img src="https://godoc.org/github.com/RTradeLtd/Temporal?status.svg"
       alt="GoDocs available" />
  </a>

  <a href="https://travis-ci.com/RTradeLtd/Temporal">
    <img src="https://travis-ci.com/RTradeLtd/Temporal.svg?branch=V2"
      alt="Travis Build Status" />
  </a>

  <a href="https://github.com/RTradeLtd/Temporal/releases">
    <img src="https://img.shields.io/github/release-pre/RTradeLtd/Temporal.svg?label=preview"
      alt="Release" />
  </a>

  <a href="https://app.fossa.io/projects/git%2Bgithub.com%2FRTradeLtd%2FTemporal?ref=badge_shield" alt="FOSSA Status">
    <img src="https://app.fossa.io/api/projects/git%2Bgithub.com%2FRTradeLtd%2FTemporal.svg?type=shield"/>
  </a>

  <a href="https://goreportcard.com/report/github.com/RTradeLtd/Temporal">
    <img src="https://goreportcard.com/badge/github.com/RTradeLtd/Temporal"
      alt="Clean code" />
  </a>
</p>

<br>

## About Temporal

Temporal is an enterprise-grade storage solution featuring an easy to consume API that can be easily integrated into your existing application stack, providing all the benefits of the distributed web, without any of the overhead that comes with running distributed storage nodes. Currently supporting public and private IPFS+IPNS usage, subsequent releases will bring about support for additional protocols like STORJ, and SWARM.

Temporal's API comes in two flavours, hosted or on-site. Should you not have the resources, or interest in maintaining your own infrastructure you can take advantage of our hosted API running in our very own datacenter. Those which have the interest, and/or resources may deploy Temporal within your own environments. For those that deploy Temporal themselves, we offer paid for support, installation, tutorials, and product usage information sessions allowing organizations to leverage all the capabilities that Temporal offers.

Temporal is modular such that the underlying protocols it connects to, can easily be upgraded, and replaced with without having to change the overall architecture. Simply write the interface for whatever protocol you're interested in, and a subsequent RabbitMQ consumer, and you're good to go. For a guideline on how this can be accomplished see [protocol-expansion.md](/docs/protocol-expansion.md)

We have a [telegram chat](https://t.me/RTradeTEMPORAL), feel free to join and ask any questions you may have that could not be answered by reading the documentation here. We have an in-depth [wiki](https://rtradetechnologies.atlassian.net/wiki/spaces/TEM/overview) which contains additional information such as architectural diagrams, design decisions, and more.

### Goals

* Provide an easy to use interface into distributed and decentralized storage technologies.
* Target developers via the API, and non-developers via the web interface.
* Educate about decentralized and distributed storage technologies
* Introduce these new storage technologies to audiences who may have otherwise not heard of them
* Help organizations make informed decisions about whether or not integrating distributed and decentralized storage technologies is the right choice for your business needs.

## Versioning Policy

Information about our versioning policies are available in [VERSIONING.md](/VERSIONING.md)

## Web Interfaces

For those less interested in API usage, we have a web interface which can be used with two methods of access:

* [Clearnet](https://temporal.cloud) (recommended)
* [I2P](http://riqdsr6ijsujw4tagdufhbv7drlghe2cljy2xow3irvy7grq34fq.b32.i2p/)

Please note that support for the I2P Interface is very experimental at the moment and does not offer HTTPS, as well as very infrequent updates.There are no stability, or even functionality guarantees for the I2P interface.

## How We're Different

We aren't doing an ICO,  and we're not wasting our development efforts on redesigning the wheel with some new fangled storage protocol, and blockchain solution. Although we're using bleeding edge technology, we're commited to using names, and open source software that is already tested, and that has a thriving development community behind them. And finally, results matter; It is far to common in this space for companies to ask you to hand over your hand earned cash on the fleeting promise that it will lead to something, but that something is either never delivered, or extremely lack in features, and is not the original idea which was sold.

We aren't running in a third-party cloud environment, as our hosted API runs in our very own datacenter. The only third-party cloud environments we utilize are for off-site backups, and service should any critical infrastructure failures occur in our datacenter. In the future as we expand and receive more funding, we will build-out another facility to provide our off-site backups and service endpoints. We are committed to bringing back control to users data, and we believe the only way to fully do this, is to build our own infrastructure, for which we maintain all control, and design decisions for. As a startup, we have to rely on third-party cloud providers for backups and off-site service points due to funding limitations, however we realize this is undesirable. This is one of our main concerns, and is something we are intent on resolving when our company grows, and we receive more funding.

## Funding

Currently the project is paid for by RTrade Technologies Ltd, and we will *not* be doing an ICO. Funding is derived from private investment, mining farm profits, as well as purchasing of RTC.

Should you wish to consider donations, or private investment email admin@rtradetechnologies.com.

Should you wish to contribute not just to TEMPORAL, but to the overall success of RTrade and our platform, you may purchase RTC for ETH from our [RTCETH Smart Contract](https://etherscan.io/address/0x40e68e3F58b9C1928954BEe5dEcC09A45aA531f8#code)

## Media

Channels:

* [Medium](https://medium.com/@rtradetech)
* [Twitter](https://twitter.com/RTradeTech)
* [Steemit](https://steemit.com/@rtrade)
* [LinkedIn](https://www.linkedin.com/company/rtrade-technologies/)

## Data Privacy

Our datacenter and cloud environments are all located within Canada, which has exceptional data-privacy laws. We comply with all laws and regulations surrounding data storage regulation within Canada. Should you feel like there is any discrepancy here, please contact us at admin@rtradetechnologies.com and we will be happy to resolve your concerns, and if there's anything we need to change, we will do so.

One of the big concerns with IPFS, and even cloud data storage in general is encryption. As IPFS doesn't yet support native data encryption, we allow users to encrypt their data using AES256-CFB. While this is better than storing data without encryption on IPFS, there are still some concerns with encrypted data storage on IPFS. Namely, if anyone is ever able to discover the content hash and pin the content, it will always be available to them. This is of big concern when using encryption algorithms as it is theoretically possible for someone to persist that data within their own storage system until the desired ciphers are broken, and they can crack the encryption algorithm. If this is something that is of concern to you, and even better solution is to encrypt your data, and store in on private networks. We have plans to eventually migrate to AES256-GCM which is more secure than AES256-CFB, and allow encryption of data with IPFS keys.

## Usage and Features

### Spinning up a Node

Once you have a `config.json` set up (a template can be generated using `temporal init`), you can run the following commands to use docker-compose to spin up Temporal:

```shell
$> curl https://raw.githubusercontent.com/RTradeLtd/Temporal/V2/temporal.yml --output temporal.yml
$> docker-compose -f temporal.yml up
```

Refer to the `temporal.yml` documentation for more details.

### API Documentation

Our API documentation is through [postman](https://documenter.getpostman.com/view/4295780/RWEcQM6W#intro)

### Features

* API based
* Detailed logging
* Open source
* Public+Private IPFS usage
* Public+Private IPNS usage
* Private IPFS Network Management
* Modular design allowing for ease of integration with multiple storage protocols
* Optional data encryption before your content leaves server memory, and touches any distributed storage network.
* Redundant architecture designed for running two of every service, allowing for service availability despite catastrophic hardware failures.

### Supported Technologies

Currently we fully support all non-experimental IPFS and IPNS feature-sets. Features like UnixFS, MFS are on-hold until their specs, and implementations become more stable and usable within production environments. Additional protocols like STORJ, and SWARM will be added, fully supporting public+private integrations. At the moment, the next planned protocol is STORJ, with alpha integration expected near the of January/February 2019.

### System Monitoring

Temporal is designed to be monitored witha combination of Zabbix, and Grafana+Prometheus. Zabbix is used for operating system, and hardware level metric collection, while Grafana+Prometheus are used to scrape API metrics, along with IPFS and IPFS Cluster node metrics. We include Zabbix templates, and Grafana graphs within the `setup/configs` folder.

## Development

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

### Vendoring

To update or rebuild the dependencies, run:

```bash
$ make vendor
```

This requires the GX dependency management tool to be installed.

### Repository Contents

Refer to the [documentation](https://godoc.org/github.com/RTradeLtd/Temporal) for an overview of Temporal's package structure.

The project also depends on a [range of sub-repositories](https://github.com/search?q=topic%3Atemporal+org%3ARTradeLtd&type=Repositories).

## License

Temporal is licensed under Apache-2, with the bulk of our sub-projects also licensed under Apache-2. However certain subprojects, particularly those which act as convenient wrappers, like our [rfts](https://github.com/RTradeLtd/rtfs) are licensed under MIT.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FRTradeLtd%2FTemporal.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FRTradeLtd%2FTemporal?ref=badge_large)

## Thanks

Without open source, Temporal wouldn't be possible, as such we would like to extend our thanks to all of the open source projects for which Temporal depends on. If you notice any are missing from the list below please open an issue and we will add it to the list:

* https://github.com/ipfs
* https://github.com/miguelmota/go-solidity-sha3
* https://github.com/libp2p
* https://github.com/ethereum/
* https://github.com/jinzhu/gorm
* https://github.com/gin-gonic/gin
* https://github.com/minio/minio
* https://github.com/streadway/amqp
