package tcp

import (
	"context"
	"time"

	manet "gx/ipfs/QmQVUtnrNGtCRkCMpXgpApfzQjc8FDaDVxHqWH8cnZQeh5/go-multiaddr-net"
	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
	mafmt "gx/ipfs/QmXZVrmSiWPyuET6BMbGK3PjzVnQBrcnYc4GkHKhK2KXLp/mafmt"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	rtpt "gx/ipfs/QmZJYYsGSNSFMAAQ2bxtdmy8SuYWVrBVH84gcpoDUwQvDF/go-reuseport-transport"
	tpt "gx/ipfs/Qmb3qartY8DSgRaBA3Go4EEjY1ZbXhCcvmc4orsBKMjgRg/go-libp2p-transport"
	tptu "gx/ipfs/Qmc9KUyhx1adPnHX2TBjEWvKej2Gg2kvisAFoQ74UiWYhd/go-libp2p-transport-upgrader"
	logging "gx/ipfs/QmcuXC5cxs79ro2cUuHs4HQ2bkDLJUYokwL8aivcX6HW3C/go-log"
)

// DefaultConnectTimeout is the (default) maximum amount of time the TCP
// transport will spend on the initial TCP connect before giving up.
var DefaultConnectTimeout = 5 * time.Second

var log = logging.Logger("tcp-tpt")

// TcpTransport is the TCP transport.
type TcpTransport struct {
	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	Upgrader *tptu.Upgrader

	// Explicitly disable reuseport.
	DisableReuseport bool

	// TCP connect timeout
	ConnectTimeout time.Duration

	reuse rtpt.Transport
}

var _ tpt.Transport = &TcpTransport{}

// NewTCPTransport creates a tcp transport object that tracks dialers and listeners
// created. It represents an entire tcp stack (though it might not necessarily be)
func NewTCPTransport(upgrader *tptu.Upgrader) *TcpTransport {
	return &TcpTransport{Upgrader: upgrader, ConnectTimeout: DefaultConnectTimeout}
}

// CanDial returns true if this transport believes it can dial the given
// multiaddr.
func (t *TcpTransport) CanDial(addr ma.Multiaddr) bool {
	return mafmt.TCP.Matches(addr)
}

func (t *TcpTransport) maDial(ctx context.Context, raddr ma.Multiaddr) (manet.Conn, error) {
	// Apply the deadline iff applicable
	if t.ConnectTimeout > 0 {
		deadline := time.Now().Add(t.ConnectTimeout)
		if d, ok := ctx.Deadline(); !ok || deadline.Before(d) {
			var cancel func()
			ctx, cancel = context.WithDeadline(ctx, deadline)
			defer cancel()
		}
	}

	if t.UseReuseport() {
		return t.reuse.DialContext(ctx, raddr)
	}
	var d manet.Dialer
	return d.DialContext(ctx, raddr)
}

// Dial dials the peer at the remote address.
func (t *TcpTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (tpt.Conn, error) {
	conn, err := t.maDial(ctx, raddr)
	if err != nil {
		return nil, err
	}
	return t.Upgrader.UpgradeOutbound(ctx, t, conn, p)
}

// UseReuseport returns true if reuseport is enabled and available.
func (t *TcpTransport) UseReuseport() bool {
	return !t.DisableReuseport && ReuseportIsAvailable()
}

func (t *TcpTransport) maListen(laddr ma.Multiaddr) (manet.Listener, error) {
	if t.UseReuseport() {
		return t.reuse.Listen(laddr)
	}
	return manet.Listen(laddr)
}

// Listen listens on the given multiaddr.
func (t *TcpTransport) Listen(laddr ma.Multiaddr) (tpt.Listener, error) {
	list, err := t.maListen(laddr)
	if err != nil {
		return nil, err
	}
	return t.Upgrader.UpgradeListener(t, list), nil
}

// Protocols returns the list of terminal protocols this transport can dial.
func (t *TcpTransport) Protocols() []int {
	return []int{ma.P_TCP}
}

// Proxy always returns false for the TCP transport.
func (t *TcpTransport) Proxy() bool {
	return false
}

func (t *TcpTransport) String() string {
	return "TCP"
}
