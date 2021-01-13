The hosted service for this project is being shutdown. For more information [consult this github issue](https://github.com/RTradeLtd/Temporal/issues/489)

<h1 align="center">Temporal ☄️</h1>

<p align="center">
  <a href="#about-temporal"><strong>About Temporal</strong></a> · 
  <a href="#web-interfaces-and-endpoints"><strong>Web Interfaces & Endpoints</strong></a> · 
  <a href="#usage-and-features"><strong>Usage and Features</strong></a> · 
  <a href="/CONTRIBUTING.md"><strong>Contributing</strong></a> · 
  <a href="#license"><strong>License</strong></a> · 
  <a href="#thanks"><strong>Thanks</strong></a>

</p>

<p align="center">
  <a href="https://t.me/RTradeTEMPORAL">
    <img src="https://patrolavia.github.io/telegram-badge/chat.png"/>
  </a>
  <a href="https://godoc.org/github.com/RTradeLtd/Temporal">
    <img src="https://godoc.org/github.com/RTradeLtd/Temporal?status.svg"
       alt="GoDocs available" />
  </a>

  <a href="https://travis-ci.com/RTradeLtd/Temporal">
    <img src="https://travis-ci.com/RTradeLtd/Temporal.svg?branch=V2"
      alt="Travis Build Status" />
  </a>

  <a href="https://github.com/RTradeLtd/Temporal/releases">
    <img src="https://img.shields.io/github/release-pre/RTradeLtd/Temporal.svg"
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


## Multi-Language

[![](https://img.shields.io/badge/Lang-English-blue.svg)](README.md)  [![jaywcjlove/sb](https://jaywcjlove.github.io/sb/lang/chinese.svg)](README-zh.md)


## About Temporal

Temporal is an enterprise-grade storage solution that allows you easily integrate with distributed storage technologies like IPFS, without sacrificing functionality with an easy to use API leveraging all the benefits the distributed web has to offer.

Temporal's API comes in two flavors, hosted or on-site. Should you not have the resources, or interest in maintaining your own infrastructure you can take advantage of our hosted API running in our very own datacenter. Those who have the interest, and/or resources may deploy Temporal within your own environments. For those that deploy Temporal themselves, we offer paid for support, installation, tutorials, and product usage information sessions allowing organizations to leverage all the capabilities that Temporal offers.

Temporal is modular such that the underlying protocols it connects to, can easily be upgraded, and replaced with without having to change the overall architecture. See our [protocol-expansion.md](/docs/protocol-expansion.md) documentation for details on extending the available functionality.

We have comprehensive API documentation available [here](https://gateway.temporal.cloud/ipns/docs.api.temporal.cloud) as well as an in-depth [wiki](https://rtradetechnologies.atlassian.net/wiki/spaces/TEM/overview) which contains additional information such as architectural diagrams, design decisions, and more.

### Benefits Over Direct IPFS HTTP API Access

* Authenticated API access
* Better performance via load balanced requests to different ipfs nodes
* Leverage multiple different nodes from a single API
  * Requests are distributed round-robin across different nodes
* Built-in handling of IPFS and IPFS Cluster requests
* Content encryption
* At rest encrypted of IPFS keys
* More efficient IPNS publishing
* Easy to scale up, simply spin up additional queue runners with a single command

### Goals

* Provide an easy to use interface into distributed and decentralized storage technologies.
* Target developers via the API, and non-developers via the web interface.
* Educate about decentralized and distributed storage technologies.
* Introduce these new storage technologies to audiences who may have otherwise not heard of them.
* Help organizations make informed decisions about whether or not integrating distributed and decentralized storage technologies is the right choice for your business needs.

## Versioning Policy

Information about our versioning policies are available in [VERSIONING.md](/VERSIONING.md)

## Web Interfaces And Endpoints

We have a variety of methods to access the variety of APIs and web interfaces we expose to use IPFS, via the clearnet, and as hidden services via I2P. For those that want maximal privacy we recommend using our I2P hidden services, however they require running an I2P router to access. If you want a cheap way of running an always online I2P router you can use a Raspberry Pi [and use this script to install the I2P router, it's what the CTO for RTrade does and it works quite well](https://github.com/bonedaddy/rpi-utils/blob/b780fa8d1cbb4c822b4baa0c98520dd8cf1d33ff/scripts/install_i2p.sh).

### Clearnet

* [Web Interaface](https://play2.temporal.cloud)
* [IPFS HTTP API](https://api.ipfs.temporal.cloud/ping)
* [Temporal API](https://api.temporal.cloud/v2/systems/check)

### I2P

The I2P endpoints are the same production endpoints that you use when accessing our platform via the clearnet, and as such you can freely use the I2P endpoints with the accounts you use on the clearnet. 

* [Web Interface](http://7ds4k3ua5zzmbnfq4pvqojtemornhzyakwagcvedyycdc3ihkvja.b32.i2p)
* [IPFS HTTP API](http://tunsbg26fowkptk3j3lgg3uzwei4disqfygfjvkjidnvparpbfza.b32.i2p/ping)
* [Temporal API](http://7t4iy4eopl4w2dbqj6atrn2gtxirdwqhni4jny2pqbaictjhjcjq.b32.i2p/v2/systems/check)

Please note that our I2P endpoints may suffer availability issues as they are not yet implemented with high availability, and statically point to one of your backends. Additionally if our main service location suffers an outage redirective traffic to our offsite location, the I2P services will be offline. High availability for our I2P endpoints is a work in progress.

For more information on the I2P infrastructure, [please consult this blog post](https://medium.com/temporal-cloud/temporal-i2p-infrastructure-refresh-52fe40751f8b).

## How We're Different

Temporal gives everyone the chances to run their own enterprise grade IPFS infrastructure without needing to rely on third-party API providers that dominate the blockchain development market. However if you are unable to do this, you can use our hosted API running in our very own datacenter, using the exact same open-source code you see here. 

By using our hosted API you can experience the same level of enterprise quality service that you can setup in your own environments, whether that be cloud VMs, your own datacenter, or a server under your bed.  Additionally, by building your applications decentralized or centralized with Temporal, you won't be vendor locked in if you decide to transition to a self-hosted infrastructure, because all it takes is changing a single URL that your application points to.

Temporal provides a first of its kind API that outweighs every other third-party IPFS and decentralized storage API on the market so you can fully leverage all the benefits the distributed web has to offer.

## Funding

Currently the project is paid for by RTrade Technologies Ltd, and we will *not* be doing an ICO. Funding is derived from private investment, mining farm profits, as well as purchasing of RTC.

Should you wish to consider donations, or private investment, email admin@rtradetechnologies.com.

Should you wish to contribute not just to Temporal, but to the overall success of RTrade and our platform, you may purchase RTC for ETH from our [RTCETH Smart Contract](https://etherscan.io/address/0x40e68e3F58b9C1928954BEe5dEcC09A45aA531f8#code)

## Media

Channels:

* [Medium](https://medium.com/@rtradetech)
* [Twitter](https://twitter.com/RTradeTech)
* [Steemit](https://steemit.com/@rtrade)
* [LinkedIn](https://www.linkedin.com/company/rtrade-technologies/)

Selected Content:

* [Podcast with postables discussing IPFS, and Temporal](https://www.youtube.com/watch?v=TDvgcdMxmzo&feature=youtu.be)
* [ChainLink and RTrade partnership announcement](https://steemit.com/cryptocurrency/@rtrade/rtrade-technologies-to-use-chainlink-to-provide-oracles-for-high-quality-off-chain-data-storage)

## Data Privacy

Our datacenter and cloud environments are all located within Canada, which has exceptional data-privacy laws. We comply with all laws and regulations surrounding data storage regulation within Canada. Should you feel like there is any discrepancy here, please contact us at admin@rtradetechnologies.com and we will be happy to resolve your concerns, and if there's anything we need to change, we will do so.

One of the big concerns with IPFS, and even cloud data storage in general is encryption. As IPFS doesn't yet support native data encryption, we allow users to encrypt their data using AES256-CFB. While this is better than storing data without encryption on IPFS, there are still some concerns with encrypted data storage on IPFS. Namely, if anyone is ever able to discover the content hash and pin the content, it will always be available to them. This is of big concern when using encryption algorithms as it is theoretically possible for someone to persist that data within their own storage system until the desired ciphers are broken, and they can crack the encryption algorithm. If this is something that is of concern to you, an even better solution is to encrypt your data, and store in on private networks. We have plans to eventually migrate to AES256-GCM which is more secure than AES256-CFB, and allow encryption of data with IPFS keys.

## Usage and Features

Before attempting to use Temporal you will need to install it. Even if you are going to be using our dockerized tooling, an install of Temporal is needed primarily for configuration file initialization.

Please note that a full-blown Temporal instance including the payment processing backend can take awhile, and requires an API key for [ChainRider](https://chainrider.io/) as well as a fully synced [geth node](https://github.com/ethereum/go-ethereum), and [bchd node](https://github.com/gcash/bchd). We will *not* be covering the setup of either chainrider, geth, and bchd, please consult appropriate documentation for setting those up. Should you want to read about our payment processing backedn see [RTradeLtd/Pay](https://github.com/RTradeLtd/Pay)

For a minimal setup you do not need to configure anything related to the aforementioned topics (geth, pay, bch, etc..). It is worth mentioning though that running a minimal setup doesn't disable the payment API calls, so if someone were to place these API calls against your minimal setup it might result in unexpected errors, such as panics. If this happens please open a bug report on github.

The rest of this usage documentation will be covering a bare-minimum Temporal setup which does not include any payment processing capabilities. Thus you will not be able to "purchase credits" the remedy to this is to manually alter user account balances, or promote a user to partner tier, registering an organization, and then creating all new users under that organization. This effectively side-steps the billing process, and requires no manual management of user credits. 

For details on organization management, and the entire API please consult  our [api docs](https://gateway.temporal.cloud/ipns/docs.api.temporal.cloud/account.html#organization-management).

### Installing Temporal

The first thing you need to do is install Temporal, for which there are two main options:

1) Compiling from source
2) Downloading pre-built binaries

**Compiling From Source:**

This is quite a bit more complicated, and requires things like a proper golang version installed. Unless you have specific reason to compile from source, it is recommended you skip this and stick with download the pre-built binaries.

If you do want to download and build from source, be aware the download process can take *A LONG TIME* depending on your bandwidth and internet speed. Usually it takes up to 30 minutes.

Should you still want to do this, download the [install_from_source.sh script](./setup/scripts/misc/install_from_source.sh). This will ensure you have the proper go version, download the github repository, compile the cli, and install it to `/bin/temporal`.

```bash
#! /bin/bash

# setup install variabels
GOVERSION=$(go version | awk '{print $3}' | tr -d "go" | awk -F "." '{print $2}')
WORKDIR="/tmp/temporal-workdir"
# handle golagn version detection
if [[ "$GOVERSION" -lt 11 ]]; then
    echo "[ERROR] golang is less than 1.11 and will produce errors"
    exit 1
fi
if [[ "$GOVERSION" -lt 12 ]]; then
    echo "[WARN] detected golang version is less than 1.12 and may produce errors"
fi
# create working directory
mkdir "$WORKDIR"
cd "$WORKDIR"
# download temporal
git clone https://github.com/RTradeLtd/Temporal.git
cd Temporal
# initialize submodules, and download all dependencies
make setup
# make cli binary
make install
```

**Downloading Pre-Built Binaries:**

To download pre-built binaries, currently available for Linux and Mac OSX platforms, head on over to our [github releases page](https://github.com/RTradeLtd/Temporal/releases/latest).

Download a binary for your appropriate platform, additionally you can download sha256 checksums of the released binaries for post-download integrity verification. 

You'll then want to copy the pre-built binary over to anywhere in your `PATH` environment variable.

Example for downloading and installing the Linux version:

```bash
# download the binary
wget https://github.com/RTradeLtd/Temporal/releases/download/v2.2.7/temporal-v2.2.7-linux-amd64
# download the binary checksum
wget https://github.com/RTradeLtd/Temporal/releases/download/v2.2.7/temporal-v2.2.7-linux-amd64.sha256
# store downloaded checksum output
CK_HASH=$(cat *.sha256  | awk '{print $1}')
# calculate sha256 checksum of downloaded binary
DL_HASH=$(sha256sum temporal-v2.2.7-linux-amd64 | awk '{print $1}')
# compare checksums, and copy file to `PATH` if ok
# if this doesn't show ok then your binary is corrupted or has been tampered with
if [[ "$CK_HASH" == "$DL_HASH" ]]; then  echo "ok" && sudo cp temporal-v2.2.7-linux-amd64 /usr/local/bin; fi
```

### Configuration Initialization

After downloading Temporal, regardless of your setup process you will need a configuration file. To generate a config file at `/home/doggo/config.json` run the following command:
```
temporal -config /home/doggo/config.json init
```
**Alternatively you can set the environment variable `CONFIG_DAG` and `temporal init`, along with *all other commands* will read from this location. It is recommended that you do this as it makes using the cli a lot easier**

It is **extremely** important you keep this in a directory that is only accessible to the users required to run the servivce as it contains usernames and passwords to key pieces of Temporal's infrastructure.

For an example bare-minium configuration file, check out [testenv/config.json](https://github.com/RTradeLtd/testenv/blob/master/config.json)

Note that if you did the install from source method, you will already have a config file in your home directory called `temporal-config.json`.

### Manual Setup

This exact process will vary a bit depending on the environment you are installing Temporal in. At the very least you are required to use Postgres, and RabbitMQ. The operating systems you install those, and the supplementary services on is entirely up to you, but we recommend using Ubuntu 18.04LTS. For instructions on setting up Postgres see their [documentation](https://www.postgresql.org/docs/10/tutorial-start.html). For instructions on setting up RabbitMQ consult their [documentation](https://www.rabbitmq.com/download.html). We do go into a bit of a setup process for RabbitMQ in the confluence page linked below, although it is always good to read official sources.

For the manual setup process using Ubuntu 18.04LTS consult our [confluence page](https://rtradetechnologies.atlassian.net/wiki/spaces/TEM/pages/55083603/Installing+Temporal+In+Production). For the manual setup process using other operating systems, please read the confluence page and adjust the commands as needed. The confluence page covers filling out the needed parts of the configuration file.

### Dockerized Setup

The docker-compose file defaults to placing everything in `/data/temporal`, so for this part of the tutorial we will be using that particular default. If you want to override it you can use the `BASE=/path/to/base` variable.

First off you'll need to copy the Temporal config file to `/data/temporal/config.json` then you can proceed with the rest of the steps. Ensure that the config file is pointing to a postgresql and rabbitmq docker container, or server that is reachable by the docker containers that will be started up.

Additionally you'll need to make sure that any tls certificates, and files needed by the api service are appropriately located within `/data/temporal`.

To download the docker-compose file:

```shell
$> curl https://raw.githubusercontent.com/RTradeLtd/Temporal/master/temporal.yml --output temporal.yml
```

Then afterwards to execute the docker-compose file using the latest version of the temporal docker image run

```shell
$> env TEMPORAL=latest docker-compose -f temporal.yml up
```

The standalone Temporal Docker image is available on [Docker Hub](https://hub.docker.com/r/rtradetech/temporal).

### API Documentation

Our API documentation has been redesigned to use slate, hosted through IPFS. The main way to view it is through our [gateway](https://gateway.temporal.cloud/ipns/docs.api.temporal.cloud/). However, in theory it is viewable across any gateway by navigating to `/ipns/docs.api.temporal.cloud`

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

Currently we fully support all non-experimental IPFS and IPNS feature-sets. Features like UnixFS, MFS are on-hold until their specs, and implementations become more stable and usable within production environments.

### System Monitoring

Temporal is designed to be monitored with a combination of Zabbix, and Grafana + Prometheus. Zabbix is used for operating system, and hardware level metric collection, while Grafana + Prometheus are used to scrape API metrics, along with IPFS and IPFS Cluster node metrics. We include Zabbix templates, and Grafana graphs within the `setup/configs` folder.

## License

In order to better align with the same open source values that originally inspired this project, Temporal has been changed to an MIT license for its production release. Originally, I ([postables](https://github.com/postables)), intended to release under Apache 2.0, however I think to truly help the open-source, and IPFS movement launching under the MIT license is needed.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FRTradeLtd%2FTemporal.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2FRTradeLtd%2FTemporal?ref=badge_large)

## Thanks

Without open source, Temporal wouldn't be possible, as such we would like to extend our thanks to all of the open source projects on which Temporal depends on. If you notice any are missing from the list below, please open an issue and we will add it to the list.

* https://github.com/ipfs
* https://github.com/miguelmota/go-solidity-sha3
* https://github.com/libp2p
* https://github.com/ethereum/
* https://github.com/jinzhu/gorm
* https://github.com/gin-gonic/gin
* https://github.com/streadway/amqp
* https://github.com/gcash/bchwallet
