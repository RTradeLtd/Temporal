package cidenc

import (
	"testing"

	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	mbase "gx/ipfs/QmekxXDhCxCJRNuzmHreuaT3BsuJcsjcXWNrtV9C8DRHtd/go-multibase"
)

func TestCidEncoder(t *testing.T) {
	cidv0str := "QmdfTbBqBPQ7VNxZEYEj14VmRuZBkqFbiwReogJgS1zR1n"
	cidv1str := "zdj7Wkkhxcu2rsiN6GUyHCLsSLL47kdUNfjbFqBUUhMFTZKBi"
	cidb32str := "bafybeihdwdcefgh4dqkjv67uzcmw7ojee6xedzdetojuzjevtenxquvyku"
	cidv0, _ := cid.Decode(cidv0str)
	cidv1, _ := cid.Decode(cidv1str)

	testEncode := func(enc Encoder, cid cid.Cid, expect string) {
		actual := enc.Encode(cid)
		if actual != expect {
			t.Errorf("%+v.Encode(%s): expected %s but got %s", enc, cid, expect, actual)
		}
	}

	testRecode := func(enc Encoder, cid string, expect string) {
		actual, err := enc.Recode(cid)
		if err != nil {
			t.Errorf("%+v.Recode(%s): %s", enc, cid, err)
			return
		}
		if actual != expect {
			t.Errorf("%+v.Recode(%s): expected %s but got %s", enc, cid, expect, actual)
		}
	}

	enc := Encoder{Base: mbase.MustNewEncoder(mbase.Base58BTC), Upgrade: false}
	testEncode(enc, cidv0, cidv0str)
	testEncode(enc, cidv1, cidv1str)
	testRecode(enc, cidv0str, cidv0str)
	testRecode(enc, cidv1str, cidv1str)
	testRecode(enc, cidb32str, cidv1str)

	enc = Encoder{Base: mbase.MustNewEncoder(mbase.Base58BTC), Upgrade: true}
	testEncode(enc, cidv0, cidv1str)
	testEncode(enc, cidv1, cidv1str)
	testRecode(enc, cidv0str, cidv1str)
	testRecode(enc, cidv1str, cidv1str)
	testRecode(enc, cidb32str, cidv1str)

	enc = Encoder{Base: mbase.MustNewEncoder(mbase.Base32), Upgrade: false}
	testEncode(enc, cidv0, cidv0str)
	testEncode(enc, cidv1, cidb32str)
	testRecode(enc, cidv0str, cidv0str)
	testRecode(enc, cidv1str, cidb32str)
	testRecode(enc, cidb32str, cidb32str)

	enc = Encoder{Base: mbase.MustNewEncoder(mbase.Base32), Upgrade: true}
	testEncode(enc, cidv0, cidb32str)
	testEncode(enc, cidv1, cidb32str)
	testRecode(enc, cidv0str, cidb32str)
	testRecode(enc, cidv1str, cidb32str)
	testRecode(enc, cidb32str, cidb32str)
}
