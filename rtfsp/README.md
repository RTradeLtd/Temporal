# RTFSP

This is the private implementation of Temporal'S IPFS setup, and allows for the creation, configuration, and management of private IPFS networks, while also integration of private IPFS networks with Temporal

# Architecture

* Using a PSK (Public Shared Key) infrastructure
    * key is store in `IPFS_PATH/swarm.key`
    * Uses text based multicoding with `/key/swarm/psk/1.0.0/` as a name
    * Following that is a second text based multicodec defining encoding of the following data (base16, base58, base64)
        ```
        /key/swarm/psk/1.0.0
        /b16
        FFFF
        ```


## Architecture - Swarm Config

* 