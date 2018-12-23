# Versioning ![](https://img.shields.io/github/release/RTradeLtd/Temporal.svg?) ![](https://img.shields.io/github/release-pre/RTradeLtd/Temporal.svg?label=preview)

This document outlines version policies that will be upheld with Temporal V2 and
all future Temporal releases. It is a reduced public version of an internal
document.

## Versioning Policy

Versions are denoted as `MAJOR.MINOR.PATCH`, and are updated as follows:

* `PATCH`: incremented for bug fixes in any Temporal project
* `MINOR`: incremented for feature additions in any Temporal project. When this
  occurs, all projects will receive the same version increment.
* `MAJOR`: incremented for backwards-incompatible ("breaking") changes in any
  Temporal project. When this occurs, all projects will receive the same version
  increment.

### Projects

Temporal projects are packages related to Temporal, but that aren't within the
main Temporal repository. These include:

* [Temporal](https://github.com/RTradeLtd/Temporal)
* [Pay](https://github.com/RTradeLtd/Pay)
* [rtfs](https://github.com/RTradeLtd/rtfs)
* [database](https://github.com/RTradeLtd/database)
* [grpc](https://github.com/RTradeLtd/grpc)
* [crypto](https://github.com/RTradeLtd/crypto)
* [cmd](https://github.com/RTradeLtd/cmd)
* [kaas](https://github.com/RTradeLtd/kaas)

These can be found under the [`temporal` tag](https://github.com/search?q=topic%3Atemporal+org%3ARTradeLtd&type=Repositories)
on the RTrade GitHub account.

### Prerelease Projects

We have a number of experimental pre-release projects that are currently
available. These projects are released as `v0.x.x` projects, and are not under
the same versioning and compatibility guarantees as released projects. These
projects include:

* [Lens](https://github.com/RTradeLtd/Lens)
* [ipfs-orchestrator](https://github.com/RTradeLtd/ipfs-orchestrator)

When these projects are finalized and production-ready, they will be tagged with
the same `MAJOR.MINOR` version as the released projects.

### Why Start Production At Version 2?
This is primarily my (postables/alex) "fault". For awhile I was the only developer
on Temporal, and it was still pretty unclear until recently (around September)
whether or not we would actually go ahead and pursue Temporal beyond just a hobby
project. As such I haphazardly handled versioning. Onboarding Robert allowed us
to iron out these inconsistencies and establish some solid ground rules for
handling versioning, and other code hygiene habits. As such, the current version
of Temporal can effectively be considered `V1`.

## Branching Policy

Temporal projects will have 2 branches: `master`, a stable branch which maintains
the state of production releases, and `dev`, a semi-stable branch that tracks
work-in-progress features.
