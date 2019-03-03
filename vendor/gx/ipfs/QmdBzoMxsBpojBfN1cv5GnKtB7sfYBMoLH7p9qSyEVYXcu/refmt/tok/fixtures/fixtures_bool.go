package fixtures

import (
	. "gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt/tok"
)

var sequences_Bool = []Sequence{
	{"true",
		[]Token{
			{Type: TBool, Bool: true},
		},
	},
	{"false",
		[]Token{
			{Type: TBool, Bool: false},
		},
	},
}
