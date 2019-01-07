# libp2p Daemon Control Protocol

The libp2p daemon is a standalone binary meant to make it easy to bring
peer-to-peer networking to new languages without fully porting libp2p and all
of its complexities.

_At the moment, this is a living document. As such, it will be susceptible to
changes until stabilization._

## Structure

### Overview

There are two pieces to the libp2p daemon:

- __Daemon__: A golang daemon that manages libp2p hosts and proxies streams to
  the end user.
- __Client__: A library written in any language that controls the daemon over
  a protocol specified in this document, allowing end users to enjoy the
  benefits of peer-to-peer networking without implementing a full libp2p stack.

### Technical Details

The libp2p daemon and client will communicate with each other over a simple Unix
socket based protocol built with [protobuf](https://developers.google.com/protocol-buffers/).

Future implementations may attempt to take advantage of shared memory (shmem)
or other IPC constructs.

## Protocol Specification

### Data Types

The data structures are defined in [pb/p2pd.proto](../pb/p2pd.proto). All messages
are varint-delimited.

### Protocol Requests

*Protocols described in pseudo-go. Items of the form [item, ...] are lists of
many items.*

#### Errors

Any response that may be an error, will take the form of:

```
Response{
  Type: ERROR,
  ErrorResponse: {
    Msg: <error message>,
  },
}
```

#### `Identify`

Clients issue an `Identify` request when they wish to determine the peer ID and
listen addresses of the daemon.

**Client**
```
Request{
  Type: IDENTIFY,
}
```

**Daemon**
```
Response{
  Type: OK,
  IdentifyResponse: {
      Id: <daemon peer id>,
      Addrs: [<daemon listen addr>, ...],
  },
}
```

#### `Connect`

Clients issue a `Connect` request when they wish to connect to a known peer on a
given set of addresses.

**Client**
```
Request{
  Type: CONNECT,
  ConnectRequest: {
    Peer: <peer id>,
    Addrs: [<addr>, ...],
    timeout: time, // optional, in seconds
  },
}
```

**Daemon**
*May return an error.*
```
Response{
  Type: OK,
}
```

#### `Disconnect`

Clients issue a `Disconnect` request when they wish to disconnect from a peer

**Client**
```
Request{
  Type: DISCONNECT,
  DisconnectRequest: {
    Peer: <peer id>,
  },
}
```

**Daemon**
*May return an error.*
```
Response{
  Type: OK,
}
```


#### `StreamOpen`

Clients issue a `StreamOpen` request when they wish to initiate an outbound
stream to a peer on one of a set of protocols.

**Client**
```
Request{
  Type: STREAM_OPEN,
  StreamOpenRequest: {
    Peer: <peer id>,
    Proto: [<protocol string>, ...],
    timeout: time, // optional, in seconds
  },
}
```

**Daemon**
*May return an error, short circuiting.*
```
Response{
  Type: OK,
  StreamInfo: {
    Peer: <peer id>,
    Addr: <peer address connected to>,
    Proto: <protocol we connected on>,
  },
}
```

After writing the response message to the socket, the daemon begins piping the
newly created stream to the client over the socket. Clients may read from and
write to the socket as if it were the stream. **WARNING**: Clients must be
careful not to read excess bytes from the socket when parsing the daemon
response, otherwise they risk reading into the stream output.

#### `StreamHandler` - Register

Clients issue a `StreamHandler` request to register a handler for inbound
streams on a given protocol. Prior to issuing the request, the client must be
listening to a Unix socket at the specified path.

**Client**
```
Request{
  Type: STREAM_HANDLER,
  StreamHandlerRequest: {
    Path: <path to unix socket client is listening to>,
    Proto: [<protocols to route to this handler>, ...],
  }
}
```

**Daemon**
*In the event that a stream binding already exists, this will overwrite that
stream binding with the one specified in the new request.*
```
Response{
  Type: OK,
}
```

#### `StreamHandler` - Inbound stream

When peers connect to the daemon on a protocol for which our client has a
registered handler, the daemon will connect to the client on the registered unix
socket.

**Daemon**
*Note: this message is NOT wrapped in a `Response` object.*
```
StreamInfo{
  Peer: <peer id>,
  Addr: <address of the peer>,
  Proto: <protocol stream opened on>,
}
```

After writing the `StreamInfo` message, the daemon will once again begin piping
data from the stream to the unix socket and vice-versa.
