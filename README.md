# Temporal (Heavy WIP) [![GoDoc](https://godoc.org/github.com/RTradeLtd/Temporal/api?status.svg)](https://godoc.org/github.com/RTradeLtd/Temporal/api)

Temporal is an enterprise-grade storage solution featuring an easy to consume API that can be re-used for a variety of applications, initially paid-for-pinning of IPFS content, and a personal cloud storage solution backed by IPFS. Soon after the MVP functionality will be expanded to connect into the Ethereum swarm network, and private IPFS swarms/clusters. Followed up plugging into STORJ and SIA. Ultimately the final product will be an enterprise ready API to allow organizations to use these distributed storage protocols in their own products. This API will be backed by a hosted version, allowing RTrade to take care of all the hard work in maintaining the necessary infrastructure, while you get to make beautiful applications using our API, which abstracts away all the nitty gritty details, and complications that come from using the new generation of storage protocols.


Technologies that will be supported by our cloud storage, and API:
* Public IPFS (current)
* Private IPFS (planned)
* Ethereum Swarm (planned)
* STORJ (planned)
* SIA (planned)

# Hosted API 

One of the main features of temporal we offer is hosted access to all the tools, and benefits that come with temporal, without having to run any of the services yourself.
While the API is easy to use and consume, there still comes work with maintaining the underlying infrastructure which is not easy, and may not be something that is suitable for you to do an on-premise deployment of the infrastructure if you don't have the resources to support it. That is where the hosted API comes into place. 

Your API infrastructure will be run on dedicated machines connected to our storage backend. At the moment since we don't support private IPFS integration with temporal currently, hosted API access for your own personal private IPFS swarms are not yet supported for on-demand registration through our web interface. Should you wish to have your own private IPFS network and still want hosted API access, please contact us privately as this is something we can setup on a case-by-case basis, but is not yet ready for automated registration.

# On Premise Deployment

There is no support offered for on-premise deployments, save for bugs that may be found in the temporal software suite itself in which case an issue may be opened up on the repository. We offer paid support for on-premise temporal deployments, as well as training to use the software to its fullest capabilities. If this is of interest to you, please contact us privately.

# Early Access Hosted API Alpha

An early access version of the hosted temporal API is online, letting you store files onto IPFS, and integrate IPFS storage into your own stack without having to run nodes. Email postables@rtradetechnologies.com for access.

# Goals

* Provide an easy to use interface with IPFS for both personal and business purposes.
* Introduce IPFS to people who may have otherwise not been able to use the technology
* Educate about decentralized and distributed storage technologies
* Provide an easy-to-use API for different decentralized and distributed storage technologies


# Data Privacy

This is a huge issue and concern for any form of cloud storage. But is seldom mentioned by any of the projects promising revolutionary storage systems, global in scope. They'll talk about all the technical benefits, and features of their storage solutions which use IPFS, but data privacy isn't discussed. Data privacy is an extreme focus of RTrade, and we will not be releasing Temporal until we are absolutely certain that all data privacy laws we need to abide by, are met. This is currently one of our primary areas of focus.

# How're We Different

We aren't doing an ICO,  and we're not wasting our development efforts on redesigning the wheel with some new fangled storage protocol, and blockchain solution. Although we're using bleeding edge technology, we're commited to using names, and open source software that is already tested, and that has a thriving development community behind them. And finally, results matter; It is far to common in this space for companies to ask you to hand over your hand earned cash on the fleeting promise that it will lead to something, but that something is either never delivered, or extremely lack in features, and is not the original idea which was sold.

# Contributing Code

If you wish to contribute, create a new branch to do your work on, and submit a pull request to V2 branch.
The only requirement is that all code submitted must include tests.

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

# Usage (to do)

## Usage: API

* `/api/v1/login`
    * This is used to login, and authenticate with temporal

* `/api/v1/register`
    * This is used to register a user account with temporal

* `/api/v1/ipfs/pin/:hash`
    * This is used to pin content to temporal
    * Note: this only pins to a local node, but will trigger a cluster wide pin of the object after it is pinned to the local node

* `/api/v1/ipfs/add-file`
    * This is used to upload a file to IPFS through temporal
    * Note: this adds it first to a local node, followed by a cluster pin 

* `/api/v1/ipfs/remove-pin/:hash`
    * This is used to remove a pin from the local ipfs node
    * Note: only the admin of the temporal instance can call this route

* `/api/v1/ipfs-cluster/pin/:hash`
    * This is used to pin a hash to the cluster

* `/api/v1/ipfs-cluster/sync-errors-local`
    * This is used to parse through a local node cluster status, and sync all errors

* `/api/v1/ipfs-cluster/remove-pin/:hash`
    * This is used to remove a pin from the cluster
    * Note: this is onyl usable by the admin of a temporal instance

## Usage: Deploying Temporal
Authentication is done with JWT, for testing with postman see this link https://medium.com/@codebyjeff/using-postman-environment-variables-auth-tokens-ea9c4fe9d3d7


Setting up postman with the tests section:

    var data = JSON.parse(responseBody); // parses the response body
                                    // into json for us
    console.log(data);
    postman.setEnvironmentVariable("token", data.token);


1) install postgresql
2) setup database tables
3) run `scripts/temporal_install.sh`
4) Setup `config.json`

When running temporal the following services need to be ran:
    1) api
    2) queue-dpa
    3) queue-dfa
    4) ipfs-cluster-queue

Before running temporal for the first time, and after changing any of the models run `Temporal migrate`

# Tips and Tricks

Make sure you set the path to the ipfs repo, and ipfs-cluster directory in your `.bashrc` or other profile file.
