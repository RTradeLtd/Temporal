package filter

import (
	"net"
	"testing"

	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
)

func TestFilter(t *testing.T) {
	f := NewFilters()

	_, ipnet, _ := net.ParseCIDR("0.1.2.3/24")
	f.AddDialFilter(ipnet)
	filters := f.Filters()
	if len(filters) != 1 {
		t.Fatal("Expected only 1 filter")
	}
	f.Remove(filters[0])

	for _, cidr := range []string{
		"1.2.3.0/24",
		"4.3.2.1/32",
		"fd00::/8",
		"fc00::1/128",
	} {
		_, ipnet, _ := net.ParseCIDR(cidr)
		f.AddDialFilter(ipnet)
	}

	for _, blocked := range []string{
		"/ip4/1.2.3.4/tcp/123",
		"/ip4/4.3.2.1/udp/123",
		"/ip6/fd00::2/tcp/321",
		"/ip6/fc00::1/udp/321",
	} {
		maddr, err := ma.NewMultiaddr(blocked)
		if err != nil {
			t.Error(err)
		}
		if !f.AddrBlocked(maddr) {
			t.Fatalf("expected %s to be blocked", blocked)
		}
	}

	for _, notBlocked := range []string{
		"/ip4/1.2.4.1/tcp/123",
		"/ip4/4.3.2.2/udp/123",
		"/ip6/fe00::1/tcp/321",
		"/ip6/fc00::2/udp/321",
		"",
	} {
		maddr, err := ma.NewMultiaddr(notBlocked)
		if err != nil {
			t.Error(err)
		}
		if f.AddrBlocked(maddr) {
			t.Fatalf("expected %s to not be blocked", notBlocked)
		}
	}
}
