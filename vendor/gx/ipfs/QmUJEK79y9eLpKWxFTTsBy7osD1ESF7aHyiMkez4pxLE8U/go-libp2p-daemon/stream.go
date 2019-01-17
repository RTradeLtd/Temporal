package p2pd

import (
	"io"
	"net"
	"sync"

	inet "gx/ipfs/QmNgLg1NTw37iWbYPKcyK85YJ9Whs1MkPtJwhfqbNYAyKg/go-libp2p-net"
	manet "gx/ipfs/QmZcLBXKaFe8ND5YHPkJRAwmhJGrVsi1JqDZNyJ4nRK5Mj/go-multiaddr-net"
	ggio "gx/ipfs/QmdxUuburamoF6zF9qjeQC4WYcWGbWuRmdLacMEsW8ioD8/gogo-protobuf/io"
)

func (d *Daemon) doStreamPipe(c net.Conn, s inet.Stream) {
	var wg sync.WaitGroup
	wg.Add(2)

	pipe := func(dst io.WriteCloser, src io.Reader) {
		_, err := io.Copy(dst, src)
		if err != nil && err != io.EOF {
			log.Debugf("stream error: %s", err.Error())
			s.Reset()
		}
		dst.Close()
		wg.Done()
	}

	go pipe(c, s)
	go pipe(s, c)

	wg.Wait()
}

func (d *Daemon) handleStream(s inet.Stream) {
	p := s.Protocol()

	d.mx.Lock()
	maddr, ok := d.handlers[p]
	d.mx.Unlock()

	if !ok {
		log.Debugf("unexpected stream: %s", p)
		s.Reset()
		return
	}

	c, err := manet.Dial(maddr)
	if err != nil {
		log.Debugf("error dialing handler at %s: %s", maddr.String(), err.Error())
		s.Reset()
		return
	}
	defer c.Close()

	w := ggio.NewDelimitedWriter(c)
	msg := makeStreamInfo(s)
	err = w.WriteMsg(msg)
	if err != nil {
		log.Debugf("error accepting stream: %s", err.Error())
		s.Reset()
		return
	}

	d.doStreamPipe(c, s)
}
