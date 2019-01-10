# libp2p Daemon PubSub Protocol

The libp2p daemon PubSub protocol allows clients to subscribe and publish to topics using libp2p PubSub.

_At the moment, this is a living document. As such, it will be susceptible to
changes until stabilization._

## Protocol Specification

### Data Types

The data structures are defined in [pb/p2pd.proto](../pb/p2pd.proto). All messages
are varint-delimited. For the DHT queries, the relevant data types are:

- `PSRequest`
- `PSResponse`

All PubSub requests will be wrapped in a `Request` message with `Type: PUBSUB`. Most
PubSub responses from the daemon will be wrapped in a `Response` with the
`PSResponse` field populated. Some responses will be basic `Response` messages to convey whether or not there was an error.

`PSRequest` messages have a `Type` parameter that specifies the specific query
the client wishes to execute.

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

#### `GET_TOPICS`
Clients can issue a `GET_TOPICS` request to get a list of topics the node is subscribed to.

**Client**
```
Request{
  Type: PUBSUB,
  PSRequest: PSRequest{
    Type: GET_TOPICS,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
  PSResponse: PSResponse{
    Topics: [<topic>, ...],
  }
}
```

#### `LIST_PEERS`
Clients can issue a `LIST_PEERS` request to get a list of IDs of peers the node is connected to.

**Client**
```
Request{
  Type: PUBSUB,
  PSRequest: PSRequest{
    Type: LIST_PEERS,
  },
}
```

**Daemon**
*Can return an error*

```
Response{
  Type: OK,
  PSResponse: PSResponse{
    PeerIDs: [<id>, ...],
  }
}
```

#### `PUBLISH`
Clients can issue a `PUBLISH` request to publish data under a topic.

**Client**
```
Request{
  Type: PUBSUB,
  PSRequest: PSRequest{
    Type: PUBLISH,
    Topic: <topic>,
    Data: <data>,
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

#### `SUBSCRIBE`
Clients can issue a `SUBSCRIBE` request to subscribe to a certain topic.

**Client**
```
Request{
  Type: PUBSUB,
  PSRequest: PSRequest{
    Type: SUBSCRIBE,
    Topic: <topic>,
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
After an OK response, the connection becomes a stream of PSMessages from the daemon. To unsubscribe from the topic, the client closes the connection.