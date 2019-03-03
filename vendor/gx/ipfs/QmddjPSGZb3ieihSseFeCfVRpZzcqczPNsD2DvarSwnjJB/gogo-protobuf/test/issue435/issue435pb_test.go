package issue435

import (
	"testing"

	proto "gx/ipfs/QmddjPSGZb3ieihSseFeCfVRpZzcqczPNsD2DvarSwnjJB/gogo-protobuf/proto"
)

func TestNonnullableDefaults(t *testing.T) {
	m := &Message{
		NonnullableOptional: SubMessage{},
		NonnullableRepeated: []SubMessage{{}},
	}
	proto.SetDefaults(m)

	if e, a := int64(7), *m.NonnullableOptional.Value; e != a {
		t.Errorf("Default not set: want %d, got %d", e, a)
	}
	if e, a := int64(7), *m.NonnullableRepeated[0].Value; e != a {
		t.Errorf("Default not set: want %d, got %d", e, a)
	}
}
