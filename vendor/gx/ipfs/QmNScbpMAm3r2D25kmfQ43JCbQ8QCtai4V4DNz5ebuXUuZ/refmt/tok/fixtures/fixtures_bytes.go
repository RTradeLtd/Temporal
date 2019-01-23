package fixtures

import (
	. "gx/ipfs/QmNScbpMAm3r2D25kmfQ43JCbQ8QCtai4V4DNz5ebuXUuZ/refmt/tok"
)

// sequences_Number contains what it says on the tin -- but be warned:
// bytes are not representable in all formats.
//
// JSON can't clearly represent binary bytes; typically in practice transforms
// to b64 strings are used, but this is application specific territory.
var sequences_Bytes = []Sequence{
	{"short byte array",
		[]Token{
			{Type: TBytes, Bytes: []byte(`value`)}, // Note 'Length' field not used; would be redundant.
		},
	},
	{"long zero byte array",
		[]Token{
			{Type: TBytes, Bytes: make([]byte, 400)},
		},
	},
}
