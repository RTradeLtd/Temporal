package config

import (
	"fmt"

	csms "gx/ipfs/QmPcoYfH3kjZmasoVuVWrjPvdX5zwG4U8iXgEgZJD14dBf/go-conn-security-multistream"
	host "gx/ipfs/QmahxMNoNuSsgQefo9rkpcfRFmQrMN6Q99aztKXf63K7YJ/go-libp2p-host"
	security "gx/ipfs/QmbyjEKtxXmZdiByBiNhfayzEuEPPBvuD2dLpHky8cHUvy/go-conn-security"
	insecure "gx/ipfs/QmbyjEKtxXmZdiByBiNhfayzEuEPPBvuD2dLpHky8cHUvy/go-conn-security/insecure"
	peer "gx/ipfs/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"
)

// SecC is a security transport constructor
type SecC func(h host.Host) (security.Transport, error)

// MsSecC is a tuple containing a security transport constructor and a protocol
// ID.
type MsSecC struct {
	SecC
	ID string
}

var securityArgTypes = newArgTypeSet(
	hostType, networkType, peerIDType,
	privKeyType, pubKeyType, pstoreType,
)

// SecurityConstructor creates a security constructor from the passed parameter
// using reflection.
func SecurityConstructor(sec interface{}) (SecC, error) {
	// Already constructed?
	if t, ok := sec.(security.Transport); ok {
		return func(_ host.Host) (security.Transport, error) {
			return t, nil
		}, nil
	}

	ctor, err := makeConstructor(sec, securityType, securityArgTypes)
	if err != nil {
		return nil, err
	}
	return func(h host.Host) (security.Transport, error) {
		t, err := ctor(h, nil)
		if err != nil {
			return nil, err
		}
		return t.(security.Transport), nil
	}, nil
}

func makeInsecureTransport(id peer.ID) security.Transport {
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(insecure.ID, insecure.New(id))
	return secMuxer
}

func makeSecurityTransport(h host.Host, tpts []MsSecC) (security.Transport, error) {
	secMuxer := new(csms.SSMuxer)
	transportSet := make(map[string]struct{}, len(tpts))
	for _, tptC := range tpts {
		if _, ok := transportSet[tptC.ID]; ok {
			return nil, fmt.Errorf("duplicate security transport: %s", tptC.ID)
		}
		transportSet[tptC.ID] = struct{}{}
	}
	for _, tptC := range tpts {
		tpt, err := tptC.SecC(h)
		if err != nil {
			return nil, err
		}
		if _, ok := tpt.(*insecure.Transport); ok {
			return nil, fmt.Errorf("cannot construct libp2p with an insecure transport, set the Insecure config option instead")
		}
		secMuxer.AddTransport(tptC.ID, tpt)
	}
	return secMuxer, nil
}
