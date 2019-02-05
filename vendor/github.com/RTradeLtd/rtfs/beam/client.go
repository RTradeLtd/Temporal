package beam

import (
	"bytes"
	"time"

	"github.com/RTradeLtd/rtfs"
)

// Laser is used to transfer content between two different private networks
type Laser struct {
	src *rtfs.IpfsManager
	dst *rtfs.IpfsManager
}

// NewLaser creates a laser client to beam content between different ipfs networks
func NewLaser(srcURL, dstURL, token string) (*Laser, error) {
	src, err := rtfs.NewManager(srcURL, token, time.Minute*10)
	if err != nil {
		return nil, err
	}
	dst, err := rtfs.NewManager(dstURL, token, time.Minute*10)
	if err != nil {
		return nil, err
	}
	return &Laser{
		src: src,
		dst: dst,
	}, nil
}

// BeamFromSource is used to transfer content from the source network to the destination network
func (l *Laser) BeamFromSource(contentHash string) error {
	data, err := l.src.Cat(contentHash)
	if err != nil {
		return err
	}
	if _, err = l.dst.Add(bytes.NewReader(data)); err != nil {
		return err
	}
	return nil
}

// BeamFromDestination is used to transfer content from the destination network to the source network
func (l *Laser) BeamFromDestination(contentHash string) error {
	data, err := l.dst.Cat(contentHash)
	if err != nil {
		return err
	}
	if _, err = l.src.Add(bytes.NewReader(data)); err != nil {
		return err
	}
	return nil
}
