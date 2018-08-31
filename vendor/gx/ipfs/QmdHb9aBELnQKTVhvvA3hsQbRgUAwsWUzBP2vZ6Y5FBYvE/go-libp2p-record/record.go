package record

import (
	pb "gx/ipfs/QmdHb9aBELnQKTVhvvA3hsQbRgUAwsWUzBP2vZ6Y5FBYvE/go-libp2p-record/pb"
)

// MakePutRecord creates a dht record for the given key/value pair
func MakePutRecord(key string, value []byte) *pb.Record {
	record := new(pb.Record)
	record.Key = []byte(key)
	record.Value = value
	return record
}
