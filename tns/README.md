# TNS (Temporal Name Server)

TNS is an attempt at creating an "Authoritative IPNS Server" for a select set of IPNS records, particular those managed by Temporal.

"Zone Files" (the `Zone` in `types.go`) are intended to be stored publicly on IPFS as raw JSON files, intended to be easily parsed. Each zone file will be stored under the IPNS name space for the associated Zone Manager public key. By querying for the zone file, you can easily parse, and determine the latest entries for name space managed by that zone file.


# Description

When creating a zone, you create a PK (`Qm...A`) that will be the manager of your zone, and for which all zone updates will be published under. This is followed by creating a PK for the zone (`Qm...B`), and a human readable name for the zone (`example.org`).

This is then saved on IPFS as an IPLD structure with the following fields which is saved under the IPNS namespace for `Qm...A`:
```json
{
    "zone_manager": "Qm...A",
    "zone_public_key": "Qm...B",
    "name": "example.org",
    "records": "nil",
    "record_names_to_public_keys": "nil"
}
```

At this point we have our main zone `example.org` which has a "signer key" of `Qm...B` that is managed by `Qm...A`. Say you want to publish a record like `website.example.org`. In order to do so, you would create a PK (`Qm...C`) with a name of `website`.  The meta data of this must container a field `signed_by` that is the sequence number of the records signed by the Zone Manager (`Qm...A`). 

So, if this is the first record being published by the Zone Manager, then it must be a signed value of "2". Why 2? Well, because the first "update" to this zone file, was the initial "publish" under the name space for the Zone Manager. Given that this is the second update, it is the second sequence number.

Now our IPLD structure looks like this:
```json
{
    "zone_manager": "Qm...A",
    "zone_public_key": "Qm...B",
    "name": "example.org",
    "records": {
        "public_key": "Qm...C",
        "name": "website",
        "meta_data": "signed value of 2 which was signed by Qm...B, and any other data"
    },
    "record_names_to_public_keys": "website -> Qm...C""
}
```

# libp2p

Each zone file will have an associated daemon running with it, using the libp2p identity of the zone manager. Leveraging this, we can provide quick communication channels to provide information about the IPNS name space, without having to resolve the IPNS record. This is deemed okay, as we are connecting to the libp2p identity of the zone manager, who is the authorized controlling entity of this particular zone.

# TODO:

* [ ] Add SECIO For LIBP2P Connections
* [ ] Add I2P Support For Daemon (leveraging SAM bridge via [this repo](https://github.com/eyedeekay/sam3-multiaddr))