package ifconnmgr

import (
	"context"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	inet "gx/ipfs/QmNgLg1NTw37iWbYPKcyK85YJ9Whs1MkPtJwhfqbNYAyKg/go-libp2p-net"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
)

type NullConnMgr struct{}

func (_ NullConnMgr) TagPeer(peer.ID, string, int)  {}
func (_ NullConnMgr) UntagPeer(peer.ID, string)     {}
func (_ NullConnMgr) GetTagInfo(peer.ID) *TagInfo   { return &TagInfo{} }
func (_ NullConnMgr) TrimOpenConns(context.Context) {}
func (_ NullConnMgr) Notifee() inet.Notifiee        { return &cmNotifee{} }

type cmNotifee struct{}

func (nn *cmNotifee) Connected(n inet.Network, c inet.Conn)         {}
func (nn *cmNotifee) Disconnected(n inet.Network, c inet.Conn)      {}
func (nn *cmNotifee) Listen(n inet.Network, addr ma.Multiaddr)      {}
func (nn *cmNotifee) ListenClose(n inet.Network, addr ma.Multiaddr) {}
func (nn *cmNotifee) OpenedStream(inet.Network, inet.Stream)        {}
func (nn *cmNotifee) ClosedStream(inet.Network, inet.Stream)        {}
