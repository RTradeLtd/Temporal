# libp2p Daemon DHT Protocol

The libp2p daemon DHT protocol allows clients to query and announce to the
libp2p DHT.

_At the moment, this is a living document. As such, it will be susceptible to
changes until stabilization._

## Protocol Specification

### Data Types

The data structures are defined in [pb/p2pd.proto](../pb/p2pd.proto). All messages
are varint-delimited. For the DHT queries, the relevant data types are:

- `DHTRequest`
- `DHTResponse`

All DHT requests will be wrapped in a `Request` message with `Type: DHT`. Most
DHT responses from the daemon will be wrapped in a `Response` with the
`DHTResponse` field populated. Some responses will be basic `Response` messages to convey whether or not there was an error.

`DHTRequest` messages have a `Type` parameter that specifies the specific query
the client wishes to execute.

The DHT protocol supports asynchronous stream responses with arbitrary numbers
of results as well as responses that return a single value. `DHTResponse`
messages have a `Type` parameter that specifies whether a response marks the
`BEGIN`ning of a stream of messages, a `VALUE` within a stream of messages, or
the `END` of a stream of messages. Single-value responses will simply return a
single `DHTResponse` with type `VALUE`.

All `DHTRequest`s also take an optional timeout in seconds.

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

#### `FIND_PEER`
Clients can issue a `FIND_PEER` request to query the DHT for a given peer's
known addresses.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: FIND_PEER,
    Peer: <peer id>,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: VALUE,
    Peer: PeerInfo{
      Id: <peer id>,
      Addrs: [<addr>, ...],
    }
  }
}
```

#### `FIND_PEERS_CONNECTED_TO_PEER`
Clients can issue a `FIND_PEERS_CONNECTED_TO_PEER` request to query the DHT for
peers directly connected to a given peer in the DHT.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: FIND_PEERS_CONNECTED_TO_PEER,
    Peer: <peer id>,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: BEGIN,
  }
}
```

**Daemon**
*Can return any number of responses like this, including 0*

```
DHTResponse{
  Type: VALUE,
  Peer: PeerInfo{
    Id: <peer id>,
    Addrs: [<addr>, ...],
  },
}
```

**Daemon**
*Marks the end of the result stream*

```
DHTResponse{
  Type: END,
}
```

#### `FIND_PROVIDERS`
Clients can issue a `FIND_PROVIDERS` request to query the DHT for peers
that have a piece of content, identified by a CID. `FIND_PROVIDERS` optionally
include a `Count`, specifying a maximum number of results to return.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: FIND_PROVIDERS,
    Cid: <content id>,
    Count: <number of results to include>,
  },
}
```

**Daemon**
*Can return an error if the CID is invalid*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: BEGIN,
  },
}
```

**Daemon**
*Can return any number of responses like this, including 0*

```
DHTResponse{
  Type: VALUE,
  Peer: PeerInfo{
    Id: <peer id>,
    Addrs: [<addr>, ...],
  },
}
```

**Daemon**
*Marks the end of the result stream*

```
DHTResponse{
  Type: END,
}
```

#### `GET_CLOSEST_PEERS`
Clients can issue a `GET_CLOSEST_PEERS` request to query the DHT routing
table for peers that are closest to a provided key.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: GET_CLOSEST_PEERS,
    Key: <content id>,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: BEGIN,
  }
}
```

**Daemon**
*Can return any number of responses like this, including 0*

```
DHTResponse{
  Type: VALUE,
  Value: <peer id>,
}
```

**Daemon**
*Marks the end of the result stream*

```
DHTResponse{
  Type: END,
}
```

#### `GET_PUBLIC_KEY`
Clients can issue a `GET_PUBLIC_KEY` request to query the DHT routing table
for a given peer's public key.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: GET_PUBLIC_KEY,
    Peer: <peer id>,
  },
}
```

**Daemon**
*Can return an error if the peer is not found*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: VALUE,
    Value: <public key>,
  }
}
```

#### `GET_VALUE`
Clients can issue a `GET_VALUE` request to query the DHT for a value stored at
a key in the DHT.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: GET_VALUE,
    Key: <key>,
  },
}
```

**Daemon**
*Can return an error if the key is not found*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: VALUE,
    Value: <value stored at key>,
  }
}
```

#### `SEARCH_VALUE`
Clients can issue a `SEARCH_VALUE` request to query the DHT for the best/most
valid value stored at a given key. It will return a stream of values,
terminating on the best/most valid value found. After the daemon finishes its query, it will update any peers in the DHT that returned stale or bad data for
the given key with the better record.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: SEARCH_VALUE,
    Key: <key>,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: BEGIN,
  }
}
```

**Daemon**
*Can return any number of responses like this, including 0*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: VALUE,
    Value: <value at key>,
  },
}
```

**Daemon**
*Marks the end of the result stream*

```
Response{
  Type: OK,
  DHTResponse: DHTResponse{
    Type: END,
  }
}
```

#### `PUT_VALUE`
Clients can issue a `PUT_VALUE` request to write a value to a key in the DHT.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: PUT_VALUE,
    Key: <key>,
    Value: <value>,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
}
```

#### `PROVIDE`
Clients can issue a `PROVIDE` request to announce that they have data
addressed by a given CID.

**Client**
```
Request{
  Type: DHT,
  DHTRequest: DHTRequest{
    Type: PROVIDE,
    Cid: <cid>,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
}
```
