package connmgr

import (
	"context"
	"testing"
	"time"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	inet "gx/ipfs/QmNgLg1NTw37iWbYPKcyK85YJ9Whs1MkPtJwhfqbNYAyKg/go-libp2p-net"
	tu "gx/ipfs/QmNvHv84aH2qZafDuSdKJCQ1cvPZ1kmQmyD4YtzjUHuk9v/go-testutil"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
)

type tconn struct {
	inet.Conn

	peer             peer.ID
	closed           bool
	disconnectNotify func(net inet.Network, conn inet.Conn)
}

func (c *tconn) Close() error {
	c.closed = true
	if c.disconnectNotify != nil {
		c.disconnectNotify(nil, c)
	}
	return nil
}

func (c *tconn) RemotePeer() peer.ID {
	return c.peer
}

func (c *tconn) RemoteMultiaddr() ma.Multiaddr {
	addr, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/1234")
	if err != nil {
		panic("cannot create multiaddr")
	}
	return addr
}

func randConn(t *testing.T, discNotify func(inet.Network, inet.Conn)) inet.Conn {
	pid := tu.RandPeerIDFatal(t)
	return &tconn{peer: pid, disconnectNotify: discNotify}
}

func TestConnTrimming(t *testing.T) {
	cm := NewConnManager(200, 300, 0)
	not := cm.Notifee()

	var conns []inet.Conn
	for i := 0; i < 300; i++ {
		rc := randConn(t, nil)
		conns = append(conns, rc)
		not.Connected(nil, rc)
	}

	for _, c := range conns {
		if c.(*tconn).closed {
			t.Fatal("nothing should be closed yet")
		}
	}

	for i := 0; i < 100; i++ {
		cm.TagPeer(conns[i].RemotePeer(), "foo", 10)
	}

	cm.TagPeer(conns[299].RemotePeer(), "badfoo", -5)

	cm.TrimOpenConns(context.Background())

	for i := 0; i < 100; i++ {
		c := conns[i]
		if c.(*tconn).closed {
			t.Fatal("these shouldnt be closed")
		}
	}

	if !conns[299].(*tconn).closed {
		t.Fatal("conn with bad tag should have gotten closed")
	}
}

func TestConnsToClose(t *testing.T) {
	cm := NewConnManager(0, 10, 0)
	conns := cm.getConnsToClose(context.Background())
	if conns != nil {
		t.Fatal("expected no connections")
	}

	cm = NewConnManager(10, 0, 0)
	conns = cm.getConnsToClose(context.Background())
	if conns != nil {
		t.Fatal("expected no connections")
	}

	cm = NewConnManager(1, 1, 0)
	conns = cm.getConnsToClose(context.Background())
	if conns != nil {
		t.Fatal("expected no connections")
	}

	cm = NewConnManager(1, 1, time.Duration(10*time.Minute))
	not := cm.Notifee()
	for i := 0; i < 5; i++ {
		conn := randConn(t, nil)
		not.Connected(nil, conn)
	}
	conns = cm.getConnsToClose(context.Background())
	if len(conns) != 0 {
		t.Fatal("expected no connections")
	}
}

func TestGetTagInfo(t *testing.T) {
	start := time.Now()
	cm := NewConnManager(1, 1, time.Duration(10*time.Minute))
	not := cm.Notifee()
	conn := randConn(t, nil)
	not.Connected(nil, conn)
	end := time.Now()

	other := tu.RandPeerIDFatal(t)
	tag := cm.GetTagInfo(other)
	if tag != nil {
		t.Fatal("expected no tag")
	}

	tag = cm.GetTagInfo(conn.RemotePeer())
	if tag == nil {
		t.Fatal("expected tag")
	}
	if tag.FirstSeen.Before(start) || tag.FirstSeen.After(end) {
		t.Fatal("expected first seen time")
	}
	if tag.Value != 0 {
		t.Fatal("expected zero value")
	}
	if len(tag.Tags) != 0 {
		t.Fatal("expected no tags")
	}
	if len(tag.Conns) != 1 {
		t.Fatal("expected one connection")
	}
	for s, tm := range tag.Conns {
		if s != conn.RemoteMultiaddr().String() {
			t.Fatal("unexpected multiaddr")
		}
		if tm.Before(start) || tm.After(end) {
			t.Fatal("unexpected connection time")
		}
	}

	cm.TagPeer(conn.RemotePeer(), "tag", 5)
	tag = cm.GetTagInfo(conn.RemotePeer())
	if tag == nil {
		t.Fatal("expected tag")
	}
	if tag.FirstSeen.Before(start) || tag.FirstSeen.After(end) {
		t.Fatal("expected first seen time")
	}
	if tag.Value != 5 {
		t.Fatal("expected five value")
	}
	if len(tag.Tags) != 1 {
		t.Fatal("expected no tags")
	}
	for tString, v := range tag.Tags {
		if tString != "tag" || v != 5 {
			t.Fatal("expected tag value")
		}
	}
	if len(tag.Conns) != 1 {
		t.Fatal("expected one connection")
	}
	for s, tm := range tag.Conns {
		if s != conn.RemoteMultiaddr().String() {
			t.Fatal("unexpected multiaddr")
		}
		if tm.Before(start) || tm.After(end) {
			t.Fatal("unexpected connection time")
		}
	}
}

func TestTagPeerNonExistant(t *testing.T) {
	cm := NewConnManager(1, 1, time.Duration(10*time.Minute))

	id := tu.RandPeerIDFatal(t)
	cm.TagPeer(id, "test", 1)

	if len(cm.peers) != 0 {
		t.Fatal("expected zero peers")
	}
}

func TestUntagPeer(t *testing.T) {
	cm := NewConnManager(1, 1, time.Duration(10*time.Minute))
	not := cm.Notifee()
	conn := randConn(t, nil)
	not.Connected(nil, conn)
	rp := conn.RemotePeer()
	cm.TagPeer(rp, "tag", 5)
	cm.TagPeer(rp, "tag two", 5)

	id := tu.RandPeerIDFatal(t)
	cm.UntagPeer(id, "test")
	if len(cm.peers[rp].tags) != 2 {
		t.Fatal("expected tags to be uneffected")
	}

	cm.UntagPeer(conn.RemotePeer(), "test")
	if len(cm.peers[rp].tags) != 2 {
		t.Fatal("expected tags to be uneffected")
	}

	cm.UntagPeer(conn.RemotePeer(), "tag")
	if len(cm.peers[rp].tags) != 1 {
		t.Fatal("expected tag to be removed")
	}
	if cm.peers[rp].value != 5 {
		t.Fatal("expected aggreagte tag value to be 5")
	}
}

func TestGetInfo(t *testing.T) {
	start := time.Now()
	gp := time.Duration(10 * time.Minute)
	cm := NewConnManager(1, 5, gp)
	not := cm.Notifee()
	conn := randConn(t, nil)
	not.Connected(nil, conn)
	cm.TrimOpenConns(context.Background())
	end := time.Now()

	info := cm.GetInfo()
	if info.HighWater != 5 {
		t.Fatal("expected highwater to be 5")
	}
	if info.LowWater != 1 {
		t.Fatal("expected highwater to be 1")
	}
	if info.LastTrim.Before(start) || info.LastTrim.After(end) {
		t.Fatal("unexpected last trim time")
	}
	if info.GracePeriod != gp {
		t.Fatal("unexpected grace period")
	}
	if info.ConnCount != 1 {
		t.Fatal("unexpected number of connections")
	}
}

func TestDoubleConnection(t *testing.T) {
	gp := time.Duration(10 * time.Minute)
	cm := NewConnManager(1, 5, gp)
	not := cm.Notifee()
	conn := randConn(t, nil)
	not.Connected(nil, conn)
	cm.TagPeer(conn.RemotePeer(), "foo", 10)
	not.Connected(nil, conn)
	if cm.connCount != 1 {
		t.Fatal("unexpected number of connections")
	}
	if cm.peers[conn.RemotePeer()].value != 10 {
		t.Fatal("unexpected peer value")
	}
}

func TestDisconnected(t *testing.T) {
	gp := time.Duration(10 * time.Minute)
	cm := NewConnManager(1, 5, gp)
	not := cm.Notifee()
	conn := randConn(t, nil)
	not.Connected(nil, conn)
	cm.TagPeer(conn.RemotePeer(), "foo", 10)

	not.Disconnected(nil, randConn(t, nil))
	if cm.connCount != 1 {
		t.Fatal("unexpected number of connections")
	}
	if cm.peers[conn.RemotePeer()].value != 10 {
		t.Fatal("unexpected peer value")
	}

	not.Disconnected(nil, &tconn{peer: conn.RemotePeer()})
	if cm.connCount != 1 {
		t.Fatal("unexpected number of connections")
	}
	if cm.peers[conn.RemotePeer()].value != 10 {
		t.Fatal("unexpected peer value")
	}

	not.Disconnected(nil, conn)
	if cm.connCount != 0 {
		t.Fatal("unexpected number of connections")
	}
	if len(cm.peers) != 0 {
		t.Fatal("unexpected number of peers")
	}
}

// see https://github.com/libp2p/go-libp2p-connmgr/issues/23
func TestQuickBurstRespectsSilencePeriod(t *testing.T) {
	cm := NewConnManager(10, 20, 0)
	not := cm.Notifee()

	var conns []inet.Conn

	// quickly produce 30 connections (sending us above the high watermark)
	for i := 0; i < 30; i++ {
		rc := randConn(t, not.Disconnected)
		conns = append(conns, rc)
		not.Connected(nil, rc)
	}

	// wait for a few seconds
	time.Sleep(time.Second * 3)

	// only the first trim is allowed in; make sure we close at most 20 connections, not all of them.
	var closed int
	for _, c := range conns {
		if c.(*tconn).closed {
			closed++
		}
	}
	if closed > 20 {
		t.Fatalf("should have closed at most 20 connections, closed: %d", closed)
	}
	if total := closed + cm.connCount; total != 30 {
		t.Fatalf("expected closed connections + open conn count to equal 30, value: %d", total)
	}
}
