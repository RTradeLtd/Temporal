package fixtures

import (
	. "gx/ipfs/QmfWqohMtbivn5NRJvtrLzCW3EU4QmoLvVNtmvo9vbdtVA/refmt/tok"
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
