package bench

import (
	"bytes"
	"testing"

	"gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt"
	"gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt/cbor"
	"gx/ipfs/QmdBzoMxsBpojBfN1cv5GnKtB7sfYBMoLH7p9qSyEVYXcu/refmt/json"
)

var fixture_mapAlpha = map[string]interface{}{
	"B": map[string]interface{}{
		"R": map[string]interface{}{
			"R": map[string]interface{}{
				"R": map[string]interface{}{
					"R": nil,
					"M": "",
				},
				"M": "asdf",
			},
			"M": "quir",
		},
	},
	"C": map[string]interface{}{
		"N": "n",
		"M": 13,
	},
	"C2": map[string]interface{}{
		"N": "n2",
		"M": 14,
	},
	"X": 1,
	"Y": 2,
	"Z": "3",
	"W": "4",
}
var fixture_mapAlpha_json = fixture_structAlpha_json
var fixture_mapAlpha_cbor = fixture_structAlpha_cbor

func Benchmark_MapAlpha_MarshalToCborRefmt(b *testing.B) {
	var buf bytes.Buffer
	exerciseMarshaller(b,
		refmt.NewMarshaller(cbor.EncodeOptions{}, &buf), &buf,
		fixture_mapAlpha, fixture_mapAlpha_cbor,
	)
}

func Benchmark_MapAlpha_MarshalToJsonRefmt(b *testing.B) {
	var buf bytes.Buffer
	exerciseMarshaller(b,
		refmt.NewMarshaller(json.EncodeOptions{}, &buf), &buf,
		fixture_mapAlpha, fixture_mapAlpha_json,
	)
}

func Benchmark_MapAlpha_MarshalToJsonStdlib(b *testing.B) {
	exerciseStdlibJsonMarshaller(b,
		fixture_mapAlpha, fixture_mapAlpha_json,
	)
}
