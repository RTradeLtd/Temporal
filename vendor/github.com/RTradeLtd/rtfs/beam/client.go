package beam

import (
	"bytes"
	"errors"
	"time"

	"github.com/RTradeLtd/rtfs"
)

// Laser is used to transfer content between two different private networks
type Laser struct {
	net1 *rtfs.IpfsManager
	net2 *rtfs.IpfsManager
}

// NewLaser creates a laser client to beam content between different ipfs networks
func NewLaser(net1URL, net2URL string) (*Laser, error) {
	net1, err := rtfs.NewManager(net1URL, nil, time.Minute*10)
	if err != nil {
		return nil, err
	}
	net2, err := rtfs.NewManager(net2URL, nil, time.Minute*10)
	if err != nil {
		return nil, err
	}
	return &Laser{
		net1: net1,
		net2: net2,
	}, nil
}

// Beam is used to transfer content bewween two different networks
func (l *Laser) Beam(sourceNet int, contentHash string) error {
	switch sourceNet {
	case 1:
		data, err := l.net1.Cat(contentHash)
		if err != nil {
			return err
		}
		if _, err = l.net2.Add(bytes.NewReader(data)); err != nil {
			return err
		}
		return nil
	case 2:
		data, err := l.net2.Cat(contentHash)
		if err != nil {
			return err
		}
		if _, err = l.net1.Add(bytes.NewReader(data)); err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid network, must be 1, or 2")
	}
}
