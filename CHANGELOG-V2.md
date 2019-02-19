# Changelog

This document tracks changes to Temporal and its related projects for all `v2.x.x`
releases. See our [versioning policy](/VERSIONING.md) for more details.

## v2.0.5

This release is focused around putting in account restrictions, enhanced pinning functionality, and the framework for deduplicated storage billing

## api/v2

* limit free accounts to 1 month pin times
* limit non-free accounts to 2 year pin times
* prevent uploads of the same content on a per-account which would result in duplicated billing
  * instead, suggest that users extend their pin times
* add api call to extend pin time


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
