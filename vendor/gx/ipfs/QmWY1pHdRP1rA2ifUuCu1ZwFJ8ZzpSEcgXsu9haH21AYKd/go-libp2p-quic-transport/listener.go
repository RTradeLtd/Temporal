package libp2pquic

import (
	"crypto/tls"
	"net"

	ic "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	tpt "github.com/libp2p/go-libp2p-transport"
	quic "github.com/lucas-clemente/quic-go"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

var quicListenAddr = quic.ListenAddr

// A listener listens for QUIC connections.
type listener struct {
	quicListener quic.Listener
	transport    tpt.Transport

	privKey        ic.PrivKey
	localPeer      peer.ID
	localMultiaddr ma.Multiaddr
}

var _ tpt.Listener = &listener{}

func newListener(addr ma.Multiaddr, transport tpt.Transport, localPeer peer.ID, key ic.PrivKey, tlsConf *tls.Config) (tpt.Listener, error) {
	lnet, host, err := manet.DialArgs(addr)
	if err != nil {
		return nil, err
	}
	laddr, err := net.ResolveUDPAddr(lnet, host)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP(lnet, laddr)
	if err != nil {
		return nil, err
	}
	ln, err := quic.Listen(conn, tlsConf, quicConfig)
	if err != nil {
		return nil, err
	}
	localMultiaddr, err := toQuicMultiaddr(ln.Addr())
	if err != nil {
		return nil, err
	}
	return &listener{
		quicListener:   ln,
		transport:      transport,
		privKey:        key,
		localPeer:      localPeer,
		localMultiaddr: localMultiaddr,
	}, nil
}

// Accept accepts new connections.
func (l *listener) Accept() (tpt.Conn, error) {
	for {
		sess, err := l.quicListener.Accept()
		if err != nil {
			return nil, err
		}
		conn, err := l.setupConn(sess)
		if err != nil {
			sess.CloseWithError(0, err)
			continue
		}
		return conn, nil
	}
}

func (l *listener) setupConn(sess quic.Session) (tpt.Conn, error) {
	remotePubKey, err := getRemotePubKey(sess.ConnectionState().PeerCertificates)
	if err != nil {
		return nil, err
	}
	remotePeerID, err := peer.IDFromPublicKey(remotePubKey)
	if err != nil {
		return nil, err
	}
	remoteMultiaddr, err := toQuicMultiaddr(sess.RemoteAddr())
	if err != nil {
		return nil, err
	}
	return &conn{
		sess:            sess,
		transport:       l.transport,
		localPeer:       l.localPeer,
		localMultiaddr:  l.localMultiaddr,
		privKey:         l.privKey,
		remoteMultiaddr: remoteMultiaddr,
		remotePeerID:    remotePeerID,
		remotePubKey:    remotePubKey,
	}, nil
}

// Close closes the listener.
func (l *listener) Close() error {
	return l.quicListener.Close()
}

// Addr returns the address of this listener.
func (l *listener) Addr() net.Addr {
	return l.quicListener.Addr()
}

// Multiaddr returns the multiaddress of this listener.
func (l *listener) Multiaddr() ma.Multiaddr {
	return l.localMultiaddr
}
