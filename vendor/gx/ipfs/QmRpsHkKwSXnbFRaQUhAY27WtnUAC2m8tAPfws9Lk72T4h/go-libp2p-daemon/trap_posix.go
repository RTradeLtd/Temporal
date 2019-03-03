// +build !windows,!plan9,!nacl,!js

package p2pd

import (
	"os"
	"os/signal"
	"syscall"
)

func (d *Daemon) trapSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	for {
		select {
		case s := <-ch:
			switch s {
			case syscall.SIGUSR1:
				d.handleSIGUSR1()
			default:
				log.Warningf("unexpected signal %d", s)
			}
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *Daemon) handleSIGUSR1() {
	// this is the state dump signal; for now just dht routing table if present
	if d.dht != nil {
		d.dht.RoutingTable().Print()
	}
}
