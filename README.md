# Temporal (Heavy WIP) [![GoDoc](https://godoc.org/github.com/RTradeLtd/Temporal/api?status.svg)](https://godoc.org/github.com/RTradeLtd/Temporal/api)

Temporal is an enterprise-grade storage solution featuring an easy to consume API that can be easily integrated into your existing application stack, providing all the benefits of the distributed web, without any of the overhead that comes with running distributed storage nodes.  Initially targetting the public IPFS network, the next release cycle will bring support for additional protocols such as Ethereum Swarm and Private IPFS network. Temporal won't stop there and will continue to evolve as the space itself evolves. At RTrade we understand that the Blockchain and Distributed technology space evolve at an extremely quick pace so we are designing Temporal with modularity in mind so that as the space evolves we can evolve seamlessly with it.

Temporal's API comes in two flavours, hosted or on-site. Should you not have the resources to run your own distributed storage nodes and infrastructure, you can take advantage of our hosted API allowing us to manage all the storage nodes and infrastructure, so all that you have to worry about is your application and using our API; We deal with all the hardware and infrastructure failures so that you can spend your hard work focusing on releasing products, not troubleshooting infrastructure failures which drain critical development resources. If however you have the infrastructure, and technical resources you can take advantage of Temporal being open source and deploy your own Temporal infrastructure. For on-site deployments we offer special paid for installations, maintenance, upgrades, and product usage information sessions allowing you to take full advantage of the power that comes with running your own Temporal infrastructure.

Temporal is being designed with a "Plug and Play" style design such that the underlying storage protocols it connects to can be swapped in and out with other protocols at will, without having to change the overall Temporal architecture, you simply need to write the interface for whatever storage protocol you want to use and it will support it. Using your new storage protocol will be as a simple as indicating a parameter in the API call for which storage networks it should plug into. In this manner, the API itself will never have to change to support new protocols.

We have a telegram chat https://t.me/RTradeTEMPORAL, feel free to join and ask any questions you may have that could not be answered by reading the documentation here. The documentation contained is fairly detailed, as all my notes are public, however it is fairly scattered and unorganized for now.

# Smart Contracts

All smart contracts used by TEMPORAL can be viewed over here https://github.com/RTradeLtd/RT-Contracts/tree/master/TEMPORAL

# Project Features

* IPNS (100%):
    * Overtime addtional IPNS functionality will be added allowing for things like automatic record republishing, however the core functionality is done!
* IPFS (100%):
    * Overtime additional IPFS functionality will be added, however the core functionality needed to use IPFS (upload, download, etc...) is done! Both for public and private IPFS networks

# Supported Technologies

Following is a list of distributed and decentralized storage technologies that Temporal currently, or plans on supporting.


IPFS (100% complete):

    Temporal supports integration with the public IPFS network, and will evolve to support new features added to IPFS so that you will have the most optimal experience possible, and never suffer from inability to access the latest and greatest features due to an API that fails to evolve as the underlying technology evolves.

    Soon after release, support for Private IPFS networks will be integrated into Temporal, allowing you to get the same benefits of the public IPFS network, but with the data security and privacy that comes with running a private network. This is extremely useful to financial institutions, data archivers, and other industries to whom data security and privacy is one of the primary concerns when integrating with any new technology

IPNS (100% Complete):

    IPNS allows for publishing of human readable names, and immutable links to changing content. IPNS integration is an optional feature with each upload to IPFS, and will allow for creation of dnslink records on our domain. Note that for the hosted API, IPNS usage alongside of IPFS pins or file uploads will incur additional charges.

    As a special/unique feature, you can plug into Temporal's IPNS setup, and use us as a trusted signer of IPNS records, while also optionally taking care of the DNS configuration, storing the DNS record underneath our domain (rtradetechnologies.com) at a URL of your choice with a subdomain name of your choosing providing it is available. Should you desire that, you can tap into our automated configuration and have that all done automatically. Optionally, we can also build out a hosted, or on-premise automated solution for your own domain as well.

Swarm (5% complete):

    Swarm blends the power of IPFS, with the power of Blockchain and is common place in protocols like Ethereum. Temporal will provide an interface into the Ethereum mainnet Swarm protocol, allowing you to store your data onto the Ethereum blockchain. You'll also take advantage of having that particular data persistently stored on our high quality Ethereum node infrastructure.  Currently the Swarm protocol isn't fully integrated yet with Ethereum, and as such, the SWAP accounting protocol isn't fully implemented yet, which means nodes in the network aren't guaranteed to persist your data which means it could dissapear into the Ether. Until SWAP is fully integrated we hope to offer increased utilizatinon and adoption of SWARM by taking advantage of the data persistence offered by Temporal, and our high quality Ethereum infrastructure

STORJ (0% complete):

    Details TBA

SIA (0% complete):

    Details TBA

# Hosted API Early Access

We have an early access alpha of our hosted API for the public IPFS network, should you wish to test it please contact me at postables@rtradetechnologies.com

There is also a telegram room you can join https://t.me/RTradeTEMPORAL,

# System Monitoring

System monitoring is done using a combination of Zabbix, and Grafana. Zabbix is used for operating system, and hardware metrics (network utilization, resource utilization, storage utilization, etc...) while Grafana is used for API and application level information for things like IPFS node statistics, API requests, etc...

All templates, and grafana graphs, as well as necessary configuration files and scripts to replicate our system monitoring backend are included in the `configs/` directory of this repository. The only configurations we don't include are our email configurations.

# Goals

* Provide an easy to use interface into distributed and decentralized storage technologies.
* Educate about decentralized and distributed storage technologies
* Introduce these new storage technologies to audiences who may have otherwise not heard of them
* Help organizations make informed decisions about whether or not integrating distributed and decentralized storage technologies is the right thing to do for your business

# Data Privacy

This is a huge issue and concern for any form of cloud storage. But is seldom mentioned by any of the projects promising revolutionary storage systems, global in scope. They'll talk about all the technical benefits, and features of their storage solutions which use IPFS, but data privacy isn't discussed. Data privacy is an extreme focus of RTrade, and we will not be releasing Temporal until we are absolutely certain that all data privacy laws we need to abide by, are met. This is currently one of our primary areas of focus.

# How're We Different

We aren't doing an ICO,  and we're not wasting our development efforts on redesigning the wheel with some new fangled storage protocol, and blockchain solution. Although we're using bleeding edge technology, we're commited to using names, and open source software that is already tested, and that has a thriving development community behind them. And finally, results matter; It is far to common in this space for companies to ask you to hand over your hand earned cash on the fleeting promise that it will lead to something, but that something is either never delivered, or extremely lack in features, and is not the original idea which was sold.

# API Usage

Currently there is no detailed API documentation beyond what is available on godoc, or in the code comments. However, there is an up to date (as of this commit) Postman collection of all the API calls that Temporal has, and can be used to interact with a Temporal cluster.

# Repository Contents

* `api/`
    * This contains the API related code to temporal, as is the primary interface for interaction with temporal
    * all web frontends, applications, etc... use this api
* `api/middleware`
    * This is middleware used by the API and handles common functionality such as database connection parameters, rabbitmq parameters, etc...
* `bindings`
    * This is the go-ethereum bindings for the smart contracts that temporal uses
* `cli`
    * basic terminal-based cli application
* `configs`
    * parent directory contains systemd service files
* `configs/grafana`
    * JSON files for our grafana graphs
* `contracts`
    * solidity source code files for temporal smart contracts
* `database`
    * this is the package used by temporal for interacting with our database backend
* `docs`
    * contains all non-readme documentation
* `models`
    * models used by temporal
* `payments`
    * golang code for interacting with the payments smart contract
* `queue`
    * all queue related code for rabbitmq and ipfs pubsub queue
* `rtfs`
    * contains temporal code for interacting with ipfs nodes
* `rtfs_cluster`
    * contains temporal code for interacting with the ipfs cluster
* `server`
    * contains server related code for interacting with the ethereum blockchain
    * this will eventually be broken up into three seperate folders corresponding to each of the smart contracts
* `utils`
    * contains utility code used by temporal

# Funding

Currently the project is paid for out of pocket, and we will *not* be doing an ICO. Any funding received through donations, or private investment will be put to the following:

* Infrastructure
    * Servers
    * Hard drives
    * Misc infrastructure (networking, etc..)
* Code Audits
* Developer Tools
* Legal
    * Lawyers
    * Ensure we meet regulatory needs
* Media
    * Educational resources development
* Hiring additional talent for Temporal enterprise to bring project to completion

Should you wish to consider donations, or private investment email admin@rtradetechnologies.com

# Media

* https://steemit.com/blockchain/@rtrade/temporal-development-update-optimizations-community-outreach-and-more-july-6th-2018
* https://steemit.com/cryptocurrency/@rtrade/temporal-developer-update-saturday-june-30th-2018
* https://steemit.com/cryptocurrency/@rtrade/temporal-a-distributed-platform-and-storage-infrastructure-service


# Thanks

Without open source, TEMPORAL wouldn't be possible, as such we would like to extend our thanks to all of the open source projects for which TEMPORAL depends on. If you notice any are missing from the list below please open an issue and we will add it to the list:

* https://github.com/ipfs
* https://github.com/miguelmota/go-solidity-sha3
* https://github.com/libp2p
* https://github.com/ethereum/
* https://github.com/jinzhu/gorm
* https://github.com/gin-gonic/gin
* https://github.com/minio/minio
* https://github.com/streadway/amqp
