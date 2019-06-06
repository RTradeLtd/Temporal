# Changelog

This document tracks changes to Temporal and its related projects for all `v2.x.x`
releases. See our [versioning policy](/VERSIONING.md) for more details.

## v2.2.0

### api/v2

* Better handling of key deletion requests [#346](https://github.com/RTradeLtd/Temporal/pull/346)
* Free credits no longer granted when upgrading to paid tier [#346](https://github.com/RTradeLtd/Temporal/pull/346)
* Implement refactored pricing model [#346](https://github.com/RTradeLtd/Temporal/pull/346)

### api/v3

* Removed into separate repository for upcoming TemporalX [#346](https://github.com/RTradeLtd/Temporal/pull/346)

### queue

* extract `rtns` publisher into separate package [#343](https://github.com/RTradeLtd/Temporal/pull/343)

### dependencies

* Update `go-ipfs` to `0.4.21` [#343](https://github.com/RTradeLtd/Temporal/pull/343)
* Update `ipfs-cluster` to `master` [#343](https://github.com/RTradeLtd/Temporal/pull/343)
* Switch back to `jinzhu/gorm` and move to `v1.9.8` [#345](https://github.com/RTradeLtd/Temporal/pull/345)

### misc

* Remove some unused variables in `cmd/temporal` and `api/v2` [#343](https://github.com/RTradeLtd/Temporal/pull/343)

## v2.1.1

### api/v2

* Add BCH Payments [#333](https://github.com/RTradeLtd/Temporal/pull/333)
* Enable multihash slection when uploading [#342](https://github.com/RTradeLtd/Temporal/pull/342)

## ci

* Fix dockerhub builds [#338](https://github.com/RTradeLtd/Temporal/pull/338)
* Update docker compose [#339](https://github.com/RTradeLtd/Temporal/pull/339)

### misc

* Add script to interconnect RTNS publishers and IPFS nodes [#332](https://github.com/RTradeLtd/Temporal/pull/332)
* README update [#340](https://github.com/RTradeLtd/Temporal/pull/340)

## v2.1.0

### api/v2

* Private network user access and management [#328](https://github.com/RTradeLtd/Temporal/pull/328)

### api/v3

* scaffold structure, implement authentication [#326](https://github.com/RTradeLtd/Temporal/pull/326)

### dependencies

* Update `go-ipfs` to `0.4.20` [#329](https://github.com/RTradeLtd/Temporal/pull/329)
* Update `ipfs-cluster` to `0.10.1` [#330](https://github.com/RTradeLtd/Temporal/pull/330)
* Move to gomod [#330](https://github.com/RTradeLtd/Temporal/pull/330)

### ci

* All releases now include sha256 checksums [#331](https://github.com/RTradeLtd/Temporal/pull/331)

## v2.0.6

### API

* Directory upload virus scan, accounting [#322](https://github.com/RTradeLtd/Temporal/pull/322)
* Allow CORS configuration through config file [#321](https://github.com/RTradeLtd/Temporal/pull/321)
* Fix cant upload bug [#320](https://github.com/RTradeLtd/Temporal/pull/320)

### dependencies

* IPFS Cluster update to 0.10.0, database update which fixed data limit increases not being applied [#319](https://github.com/RTradeLtd/Temporal/pull/319)

## v2.0.5

### docker

* Updated makefile to include installation, and configuration of the gvisor sandboxed runtime environment [#312](https://github.com/RTradeLtd/Temporal/pull/312)
* Update IPFS Cluster to 0.9.0 [#310](https://github.com/RTradeLtd/Temporal/pull/310)
* Remove unusded docker images [#309](https://github.com/RTradeLtd/Temporal/pull/309)

### api/v2

* Allow uploading directories, mainly to assist with adding websites to IPFS [#311](https://github.com/RTradeLtd/Temporal/pull/311)
* re-add Lens functionality targetting the v2 refactor [#314](https://github.com/RTradeLtd/Temporal/pull/314)
* fix jwt logging from repeatedly appending the failed user to the included fields [#318](https://github.com/RTradeLtd/Temporal/pull/318)
* add production terms and service [#318](https://github.com/RTradeLtd/Temporal/pull/318)

### queue

* Fix usage of krab in development environments [#313](https://github.com/RTradeLtd/Temporal/pull/313)
* When publishing IPNS records, if retrieving key from priamry krab fails, attempt fallback before failing [#313](https://github.com/RTradeLtd/Temporal/pull/313)

### dependencies

* Update go-ipfs to 0.4.19 [#317](https://github.com/RTradeLtd/Temporal/pull/317)
* Update travis to use, and build with go1.12 [#317](https://github.com/RTradeLtd/Temporal/pull/317)

## v2.0.4

* relevant PRs:
  * [#305](https://github.com/RTradeLtd/Temporal/pull/305)
  * [#306](https://github.com/RTradeLtd/Temporal/pull/306)
  * [#307](https://github.com/RTradeLtd/Temporal/pull/307)
  * [#308](https://github.com/RTradeLtd/Temporal/pull/308)
  
### scripts

* miscellaneus script cleanup
* add executable permissions to all scripts

### configs

* update zabbix monitoring template
  * monitor all new services
  * include graphs, and triggers for alerting

### travis

* fix personal access token for travis deployments

### queue

* fix key generation process

### grpc clients

* add client for kaas, and allow fallback mode

### deps

* general dependency update
* update [RTradeLtd/config](https://github.com/RTradeLtd/config)

## v2.0.3

* api/v2: fix private network creation ([#304](https://github.com/RTradeLtd/Temporal/pull/304))

## v2.0.2

* queue: fix key creation queue not having all consumers process the same message ([#303](https://github.com/RTradeLtd/Temporal/pull/303))

## v2.0.1

* docs: update README for V2-specific things ([#301](https://github.com/RTradeLtd/Temporal/pull/301))
* deps: upgrade all Temporal core subprojects to `~v2.0.0` ([#301](https://github.com/RTradeLtd/Temporal/pull/301))
* deps: pin IPFS to `v0.4.18` ([#301](https://github.com/RTradeLtd/Temporal/pull/301))
