// +build windows plan9 nacl js

package p2pd

func (d *Daemon) trapSignals() {
	// TODO: define signals we want to trap on Windows, if any.
	return
}
