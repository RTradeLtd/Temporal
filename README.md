<h1 align="center">Temporal ☄️</h1>

<p align="center">
  <a href="#about-temporal"><strong>About Temporal</strong></a> · 
  <a href="#web-interfaces"><strong>Web Interfaces</strong></a> · 
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

### Goals

* Provide an easy to use interface into distributed and decentralized storage technologies.
* Target developers via the API, and non-developers via the web interface.
* Educate about decentralized and distributed storage technologies.
* Introduce these new storage technologies to audiences who may have otherwise not heard of them.
* Help organizations make informed decisions about whether or not integrating distributed and decentralized storage technologies is the right choice for your business needs.

## Versioning Policy

Information about our versioning policies are available in [VERSIONING.md](/VERSIONING.md)

## Web Interfaces

For those less interested in API usage, we have a web interface which can be used with two methods of access:

* [Clearnet](https://temporal.cloud) (recommended)
* [I2P](http://riqdsr6ijsujw4tagdufhbv7drlghe2cljy2xow3irvy7grq34fq.b32.i2p/)

Please note that support for the I2P Interface is very experimental at the moment and does not offer HTTPS, as well as very infrequent updates.There are no stability, or even functionality guarantees for the I2P interface.

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

### Spinning up a Node

Once you have a `config.json` set up (a template can be generated using `temporal init`), you can run the following commands to use docker-compose to spin up Temporal:

```shell
$> curl https://raw.githubusercontent.com/RTradeLtd/Temporal/master/temporal.yml --output temporal.yml
$> docker-compose -f temporal.yml up
```

The standalone Temporal Docker image is available on [Docker Hub](https://cloud.docker.com/u/rtradetech/repository/docker/rtradetech/temporal).

Refer to the `temporal.yml` documentation for more details.

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

Currently we fully support all non-experimental IPFS and IPNS feature-sets. Features like UnixFS, MFS are on-hold until their specs, and implementations become more stable and usable within production environments. Additional protocols like STORJ, and SWARM will be added, fully supporting public+private integrations. At the moment, the next planned protocol is STORJ, with alpha integration expected near the of January/February 2019.

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
