package testutil

import (
	"bytes"

	random "gx/ipfs/QmSJ9n2s9NUoA9D849W5jj5SJ94nMcZpj1jCgQJieiNqSt/go-random"

	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	blocksutil "gx/ipfs/QmWTtpEozefF75GPw8pfsjdK12a6hZSW4CrzeecXbsVzek/go-ipfs-blocksutil"
	"gx/ipfs/QmWoXtvgC8inqFkAATB7cp2Dax7XBi9VDvSg9RCCZufmRk/go-block-format"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	bsmsg "gx/ipfs/QmYokQouMdEuZjNQop75Bwm6ZV9CxJDcxvZHeSy4Ttzrtp/go-bitswap/message"
	"gx/ipfs/QmYokQouMdEuZjNQop75Bwm6ZV9CxJDcxvZHeSy4Ttzrtp/go-bitswap/wantlist"
)

var blockGenerator = blocksutil.NewBlockGenerator()
var prioritySeq int
var seedSeq int64

func randomBytes(n int64, seed int64) []byte {
	data := new(bytes.Buffer)
	random.WritePseudoRandomBytes(n, data, seed)
	return data.Bytes()
}

// GenerateBlocksOfSize generates a series of blocks of the given byte size
func GenerateBlocksOfSize(n int, size int64) []blocks.Block {
	generatedBlocks := make([]blocks.Block, 0, n)
	for i := 0; i < n; i++ {
		seedSeq++
		b := blocks.NewBlock(randomBytes(size, seedSeq))
		generatedBlocks = append(generatedBlocks, b)

	}
	return generatedBlocks
}

// GenerateCids produces n content identifiers.
func GenerateCids(n int) []cid.Cid {
	cids := make([]cid.Cid, 0, n)
	for i := 0; i < n; i++ {
		c := blockGenerator.Next().Cid()
		cids = append(cids, c)
	}
	return cids
}

// GenerateWantlist makes a populated wantlist.
func GenerateWantlist(n int, ses uint64) *wantlist.ThreadSafe {
	wl := wantlist.NewThreadSafe()
	for i := 0; i < n; i++ {
		prioritySeq++
		entry := wantlist.NewRefEntry(blockGenerator.Next().Cid(), prioritySeq)
		wl.AddEntry(entry, ses)
	}
	return wl
}

// GenerateMessageEntries makes fake bitswap message entries.
func GenerateMessageEntries(n int, isCancel bool) []*bsmsg.Entry {
	bsmsgs := make([]*bsmsg.Entry, 0, n)
	for i := 0; i < n; i++ {
		prioritySeq++
		msg := &bsmsg.Entry{
			Entry:  wantlist.NewRefEntry(blockGenerator.Next().Cid(), prioritySeq),
			Cancel: isCancel,
		}
		bsmsgs = append(bsmsgs, msg)
	}
	return bsmsgs
}

var peerSeq int

// GeneratePeers creates n peer ids.
func GeneratePeers(n int) []peer.ID {
	peerIds := make([]peer.ID, 0, n)
	for i := 0; i < n; i++ {
		peerSeq++
		p := peer.ID(peerSeq)
		peerIds = append(peerIds, p)
	}
	return peerIds
}

var nextSession uint64

// GenerateSessionID make a unit session identifier.
func GenerateSessionID() uint64 {
	nextSession++
	return uint64(nextSession)
}

// ContainsPeer returns true if a peer is found n a list of peers.
func ContainsPeer(peers []peer.ID, p peer.ID) bool {
	for _, n := range peers {
		if p == n {
			return true
		}
	}
	return false
}

// IndexOf returns the index of a given cid in an array of blocks
func IndexOf(blks []blocks.Block, c cid.Cid) int {
	for i, n := range blks {
		if n.Cid() == c {
			return i
		}
	}
	return -1
}

// ContainsBlock returns true if a block is found n a list of blocks
func ContainsBlock(blks []blocks.Block, block blocks.Block) bool {
	return IndexOf(blks, block.Cid()) != -1
}
