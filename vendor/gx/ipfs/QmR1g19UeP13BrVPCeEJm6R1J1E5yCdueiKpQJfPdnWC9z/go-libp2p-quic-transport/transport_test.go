package libp2pquic

import (
	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	tpt "gx/ipfs/QmS4UBXoQ5QgTJA5pc62egqa5KrQRhsDHhaFHEoGUASsxp/go-libp2p-transport"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transport", func() {
	var t tpt.Transport

	BeforeEach(func() {
		t = &transport{}
	})

	It("says if it can dial an address", func() {
		invalidAddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/1234")
		Expect(err).ToNot(HaveOccurred())
		validAddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/1234/quic")
		Expect(err).ToNot(HaveOccurred())
		Expect(t.CanDial(invalidAddr)).To(BeFalse())
		Expect(t.CanDial(validAddr)).To(BeTrue())
	})

	It("supports the QUIC protocol", func() {
		protocols := t.Protocols()
		Expect(protocols).To(HaveLen(1))
		Expect(protocols[0]).To(Equal(ma.P_QUIC))
	})
})
