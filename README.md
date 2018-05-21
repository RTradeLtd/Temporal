# Temporal [![Build Status](https://travis-ci.com/RTradeLtd/Temporal.svg?token=gDSF5EBqJK8E2W8NbsUS&branch=master)](https://travis-ci.com/RTradeLtd/Temporal)


# NOTE

It was decided to open-source Temporal sooner rather than later, to help receive feedback, constructive criticism, and to create the Temporal that everyone wants to use.
Until such time as Temporal is released, anything in this repository (docs, readme, code, etc...) is likely to change. 

The repository is loosely formatted, and any issues and notes, project tracking methods, etc... may or may not appear complete, or even tangible/understandable and at some point will change. Soon :P


Questions? Send an email to postables@rtradetechnologies.com and I'll reply as soon as I can.

# Temporal

Temporal is an enterprise-grade file hosting service backed by IPFS, Swarm, and Ethereum. Initially targeting the public IPFS network, it will be expanded to included Swarm support for the Ethereum mainnet, as well as support for private IPFS networks. Afterwards integration with SIA and STORJ will be done to provide an easy to consume API to interact with those networks.

Smart contracts are used to faciliate mangement of the temporal user base, as well as for payment of storing files. Finally, they will be used as an immutable record of files stored on our service, such that anyone may independently audit that we store the data we say we store, and for the time we say we store it for. Files can be stored in intervals of 1 month periods.

Storage technologies Temporal will support (this is always under evaluation for new technologies to be added):
* IPFS
* SIA
* STORJ
* Filecoin

# Goals

* Provide an easy to use interface with IPFS for both personal and business purposes.
* Introduce IPFS to people who may have otherwise not been able to use the technology
* Educate about decentralized and distributed storage technologies
* Provide an easy-to-use API for different decentralized and distributed storage technologies

# Data Privacy

This is a huge issue and concern for any form of cloud storage. But is seldom mentioned by any of the projects promising revolutionary storage systems, global in scope. They'll talk about all the technical benefits, and features of their storage solutions which use IPFS, but data privacy isn't discussed. Data privacy is an extreme focus of RTrade, and we will not be releasing Temporal until we are absolutely certain that all data privacy laws we need to abide by, are met. This is currently one of our primary areas of focus.

# How're We Different

We aren't doing an ICO,  and we're not wasting our development efforts on redesigning the wheel with some new fangled storage protocol, and blockchain solution. Although we're using bleeding edge technology, we're commited to using names, and open source software that is already tested, and that has a thriving development community behind them. And finally, results matter; It is far to common in this space for companies to ask you to hand over your hand earned cash on the fleeting promise that it will lead to something, but that something is either never delivered, or extremely lack in features, and is not the original idea which was sold.


# Tips and Tricks

When running IPFS, since we specify a non-standard path for the IPFS repo, you must configure the proper repo environment path BEFORE running any service depending on IPFS
