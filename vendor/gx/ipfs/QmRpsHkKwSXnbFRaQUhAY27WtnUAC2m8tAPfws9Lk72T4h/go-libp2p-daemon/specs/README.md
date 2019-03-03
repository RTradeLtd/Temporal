# libp2p Daemon Specs

The daemon specs are broken into a few main pieces:

- The [Control protocol](CONTROL.md): Governs basic client interactions such as
  adding peers, connecting to them, and opening streams.
- The [DHT subsystem](DHT.md): Governs DHT client operations.
- The [Connection Manager](CM.md): Governs the connection manager API.
