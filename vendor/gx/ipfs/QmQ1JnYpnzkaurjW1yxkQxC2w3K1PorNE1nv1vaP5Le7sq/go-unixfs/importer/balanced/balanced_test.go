package balanced

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	mrand "math/rand"
	"testing"

	h "gx/ipfs/QmQ1JnYpnzkaurjW1yxkQxC2w3K1PorNE1nv1vaP5Le7sq/go-unixfs/importer/helpers"
	uio "gx/ipfs/QmQ1JnYpnzkaurjW1yxkQxC2w3K1PorNE1nv1vaP5Le7sq/go-unixfs/io"

	u "gx/ipfs/QmNohiVssaPw3KVLZik59DBVGTSm2dGvYT9eoXt5DQ36Yz/go-ipfs-util"
	chunker "gx/ipfs/QmR4QQVkBZsZENRjYFVi8dEtPL3daZRNKk24m4r6WKJHNm/go-ipfs-chunker"
	ipld "gx/ipfs/QmRL22E4paat7ky7vx9MLpR97JHHbFPrg3ytFQw6qp1y1s/go-ipld-format"
	dag "gx/ipfs/Qmb2UEG2TAeVrEJSjqsZF7Y2he7wRDkrdt6c3bECxwZf4k/go-merkledag"
	mdtest "gx/ipfs/Qmb2UEG2TAeVrEJSjqsZF7Y2he7wRDkrdt6c3bECxwZf4k/go-merkledag/test"
)

// TODO: extract these tests and more as a generic layout test suite

func buildTestDag(ds ipld.DAGService, spl chunker.Splitter) (*dag.ProtoNode, error) {
	dbp := h.DagBuilderParams{
		Dagserv:  ds,
		Maxlinks: h.DefaultLinksPerBlock,
	}

	db, err := dbp.New(spl)
	if err != nil {
		return nil, err
	}

	nd, err := Layout(db)
	if err != nil {
		return nil, err
	}

	return nd.(*dag.ProtoNode), nil
}

func getTestDag(t *testing.T, ds ipld.DAGService, size int64, blksize int64) (*dag.ProtoNode, []byte) {
	data := make([]byte, size)
	u.NewTimeSeededRand().Read(data)
	r := bytes.NewReader(data)

	nd, err := buildTestDag(ds, chunker.NewSizeSplitter(r, blksize))
	if err != nil {
		t.Fatal(err)
	}

	return nd, data
}

//Test where calls to read are smaller than the chunk size
func TestSizeBasedSplit(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	testFileConsistency(t, 32*512, 512)
	testFileConsistency(t, 32*4096, 4096)

	// Uneven offset
	testFileConsistency(t, 31*4095, 4096)
}

func testFileConsistency(t *testing.T, nbytes int64, blksize int64) {
	ds := mdtest.Mock()
	nd, should := getTestDag(t, ds, nbytes, blksize)

	r, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	dagrArrComp(t, r, should)
}

func TestBuilderConsistency(t *testing.T) {
	testFileConsistency(t, 100000, chunker.DefaultBlockSize)
}

func TestNoChunking(t *testing.T) {
	ds := mdtest.Mock()

	nd, should := getTestDag(t, ds, 1000, 2000)
	r, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	dagrArrComp(t, r, should)
}

func TestTwoChunks(t *testing.T) {
	ds := mdtest.Mock()

	nd, should := getTestDag(t, ds, 2000, 1000)
	r, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	dagrArrComp(t, r, should)
}

func arrComp(a, b []byte) error {
	if len(a) != len(b) {
		return fmt.Errorf("arrays differ in length. %d != %d", len(a), len(b))
	}
	for i, v := range a {
		if v != b[i] {
			return fmt.Errorf("arrays differ at index: %d", i)
		}
	}
	return nil
}

func dagrArrComp(t *testing.T, r io.Reader, should []byte) {
	out, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if err := arrComp(out, should); err != nil {
		t.Fatal(err)
	}
}

func TestIndirectBlocks(t *testing.T) {
	ds := mdtest.Mock()
	dag, buf := getTestDag(t, ds, 1024*1024, 512)

	reader, err := uio.NewDagReader(context.Background(), dag, ds)
	if err != nil {
		t.Fatal(err)
	}

	out, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(out, buf) {
		t.Fatal("Not equal!")
	}
}

func TestSeekingBasic(t *testing.T) {
	nbytes := int64(10 * 1024)
	ds := mdtest.Mock()
	nd, should := getTestDag(t, ds, nbytes, 500)

	rs, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	start := int64(4000)
	n, err := rs.Seek(start, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	if n != start {
		t.Fatal("Failed to seek to correct offset")
	}

	dagrArrComp(t, rs, should[start:])
}

func TestSeekToBegin(t *testing.T) {
	ds := mdtest.Mock()
	nd, should := getTestDag(t, ds, 10*1024, 500)

	rs, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	n, err := io.CopyN(ioutil.Discard, rs, 1024*4)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4096 {
		t.Fatal("Copy didnt copy enough bytes")
	}

	seeked, err := rs.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	if seeked != 0 {
		t.Fatal("Failed to seek to beginning")
	}

	dagrArrComp(t, rs, should)
}

func TestSeekToAlmostBegin(t *testing.T) {
	ds := mdtest.Mock()
	nd, should := getTestDag(t, ds, 10*1024, 500)

	rs, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	n, err := io.CopyN(ioutil.Discard, rs, 1024*4)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4096 {
		t.Fatal("Copy didnt copy enough bytes")
	}

	seeked, err := rs.Seek(1, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	if seeked != 1 {
		t.Fatal("Failed to seek to almost beginning")
	}

	dagrArrComp(t, rs, should[1:])
}

func TestSeekEnd(t *testing.T) {
	nbytes := int64(50 * 1024)
	ds := mdtest.Mock()
	nd, _ := getTestDag(t, ds, nbytes, 500)

	rs, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	seeked, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if seeked != nbytes {
		t.Fatal("Failed to seek to end")
	}
}

func TestSeekEndSingleBlockFile(t *testing.T) {
	nbytes := int64(100)
	ds := mdtest.Mock()
	nd, _ := getTestDag(t, ds, nbytes, 5000)

	rs, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	seeked, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if seeked != nbytes {
		t.Fatal("Failed to seek to end")
	}
}

func TestSeekingStress(t *testing.T) {
	nbytes := int64(1024 * 1024)
	ds := mdtest.Mock()
	nd, should := getTestDag(t, ds, nbytes, 1000)

	rs, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	testbuf := make([]byte, nbytes)
	for i := 0; i < 50; i++ {
		offset := mrand.Intn(int(nbytes))
		l := int(nbytes) - offset
		n, err := rs.Seek(int64(offset), io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
		if n != int64(offset) {
			t.Fatal("Seek failed to move to correct position")
		}

		nread, err := rs.Read(testbuf[:l])
		if err != nil {
			t.Fatal(err)
		}
		if nread != l {
			t.Fatal("Failed to read enough bytes")
		}

		err = arrComp(testbuf[:l], should[offset:offset+l])
		if err != nil {
			t.Fatal(err)
		}
	}

}

func TestSeekingConsistency(t *testing.T) {
	nbytes := int64(128 * 1024)
	ds := mdtest.Mock()
	nd, should := getTestDag(t, ds, nbytes, 500)

	rs, err := uio.NewDagReader(context.Background(), nd, ds)
	if err != nil {
		t.Fatal(err)
	}

	out := make([]byte, nbytes)

	for coff := nbytes - 4096; coff >= 0; coff -= 4096 {
		t.Log(coff)
		n, err := rs.Seek(coff, io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
		if n != coff {
			t.Fatal("wasnt able to seek to the right position")
		}
		nread, err := rs.Read(out[coff : coff+4096])
		if err != nil {
			t.Fatal(err)
		}
		if nread != 4096 {
			t.Fatal("didnt read the correct number of bytes")
		}
	}

	err = arrComp(out, should)
	if err != nil {
		t.Fatal(err)
	}
}
