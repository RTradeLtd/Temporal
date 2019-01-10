# libp2p daemon roadmap

Revision: r1; 2018-09-25

Authors (a-z):
  - @bigs
  - @raulk
  - @stebalien
  - @vyzo

## Introduction

The *libp2p daemon* provides a standalone deployment of a libp2p host, running
in its own OS process and installing a set of virtual endpoints for co-local
applications to interact with it. The daemon is written in Go and can therefore
leverage the vast ecosystem of go-libp2p modules.

By running a *single instance* of the daemon per machine, co-local applications
can leverage the libp2p stack to communicate with peers, interact with the DHT,
handle protocols, participate in pubsub, etc. no matter the language they are
developed in, nor whether a native libp2p implementation exists in that
language. Running *multiple instances* of the daemon is also possible, and
specially useful for testing purposes.

When establishing connections, the daemon handles transport selection,
security negotiation, and protocol and stream multiplexing. Streams
are mapped 1:1 to socket connections. Writes and reads to/from those
sockets are converted to writes and reads to/from the stream, allowing
any application to interact with a libp2p network through simple,
well-defined IO.

The daemon exposes a control endpoint for management, supporting basic
operations such as peer connection/disconnection, stream opening/closing, etc.
as well as operations on higher-level subsystems like the DHT, Pubsub, and
circuits.

Even though user applications can interact with the daemon in an ad-hoc manner,
we encourage the development of bindings in different languages. We envision
these as small, lightweight libraries that encapsulate the control calls and the
local transport IO, into clean, idiomatic APIs that enable the application not
only to demand actions from the daemon, but also to plug in protocol handlers
through the native constructs of that language.

A [Gerbil binding](https://github.com/vyzo/gerbil-libp2p/) has been developed,
and a Go binding is in the works.

## Short-term roadmap

These are the short-term priorities for us. If you feel something is missing,
please open a [Github issue](https://github.com/libp2p/go-libp2p-daemon/issues).

- ✅ Protobuf control API exposed over a Unix domain socket.
- ✅ Connection lifecycle: connecting and disconnecting to/from peers.
- ✅ Stream lifecycle: opening and closing streams.
- ✅ Stream <> unix socket 1:1 mapping.
- ✅ Daemon identity: auto-generated, and persisted.
- 🚧 Subsystem: DHT interactions.
- 🚧 Subsystem: Pubsub interactions.
- 🚧 Support multiaddr protocols instead of exclusively unix sockets.
- Subsystem: Circuit relay support.
- Subsystem: Peerstore operations.
- Connection notifications.
- Enabling interoperability testing between libp2p implementations.
- Go binding.
- Python binding.

## Medium-term roadmap

These are the medium-term priorities for us. If you feel something is missing,
please open a [Github issue](https://github.com/libp2p/go-libp2p-daemon/issues).

- Multi-tenancy, one application = one identity = one peer ID.
- app <> daemon isolation; trust-less scenario; programs should not be able to
  interfere or spy on streams owned by others.
- Shared-memory local transport between apps and the daemon: potentially more
  efficient than unix sockets.
- Extracting local transports as go-libp2p transports.
- Allowing "blessed" applications to act on behalf of the daemon.
- Global services implemented in the user space.
- Plugins: services providing features back to the daemon, for use by other
  tenants.
