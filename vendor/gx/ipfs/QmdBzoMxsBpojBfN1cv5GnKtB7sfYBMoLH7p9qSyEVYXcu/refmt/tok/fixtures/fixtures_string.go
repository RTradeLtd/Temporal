package fixtures

import (
	. "gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt/tok"
)

var sequences_String = []Sequence{
	{"empty string",
		[]Token{
			TokStr(""),
		},
	},
	{"flat string",
		[]Token{
			TokStr("value"),
		},
	},
	{"strings needing escape",
		[]Token{
			TokStr("str\nbroken\ttabbed"),
		},
	},
}
