package bench

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"reflect"
	"testing"

	"gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt"
)

func exerciseMarshaller(
	b *testing.B,
	subj refmt.Marshaller,
	buf *bytes.Buffer,
	val interface{},
	expect []byte,
) {
	var err error
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err = subj.Marshal(val)
	}
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(buf.Bytes(), expect) {
		panic(fmt.Errorf("result \"% x\"\nmust equal \"% x\"", buf.Bytes(), expect))
	}
}

func exerciseStdlibJsonMarshaller(
	b *testing.B,
	val interface{},
	expect []byte,
) {
	var err error
	var buf bytes.Buffer
	subj := stdjson.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err = subj.Encode(val)
	}
	if err != nil {
		panic(err)
	}
	buf.Truncate(buf.Len() - 1) // Stdlib suffixes a linebreak.
	if !bytes.Equal(buf.Bytes(), expect) {
		panic(fmt.Errorf("result \"% x\"\nmust equal \"% x\"", buf.Bytes(), expect))
	}
}

func exerciseUnmarshaller(
	b *testing.B,
	subj refmt.Unmarshaller,
	buf *bytes.Buffer,
	src []byte,
	blankFn func() interface{},
	expect interface{},
) {
	var err error
	var targ interface{}
	for i := 0; i < b.N; i++ {
		targ = blankFn()
		buf.Reset()
		buf.Write(src)
		err = subj.Unmarshal(targ)
	}
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(targ, expect) {
		panic(fmt.Errorf("result \"%#v\"\nmust equal \"%#v\"", targ, expect))
	}
}

func exerciseStdlibJsonUnmarshaller(
	b *testing.B,
	src []byte,
	blankFn func() interface{},
	expect interface{},
) {
	var err error
	var targ interface{}
	for i := 0; i < b.N; i++ {
		targ = blankFn()
		subj := stdjson.NewDecoder(bytes.NewBuffer(src))
		err = subj.Decode(targ)
	}
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(targ, expect) {
		panic(fmt.Errorf("result \"%#v\"\nmust equal \"%#v\"", targ, expect))
	}
}
