package fixtures

import (
	. "gx/ipfs/QmfWqohMtbivn5NRJvtrLzCW3EU4QmoLvVNtmvo9vbdtVA/refmt/tok"
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
