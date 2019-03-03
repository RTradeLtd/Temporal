package doc_test

import (
	"bytes"
	"fmt"

	"gx/ipfs/QmXj63M2w2Pq7mnBpcrs7Va8prmfhvfMUNqVhJ9TgjiMbT/cobra"
	"gx/ipfs/QmXj63M2w2Pq7mnBpcrs7Va8prmfhvfMUNqVhJ9TgjiMbT/cobra/doc"
)

func ExampleGenManTree() {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "my test program",
	}
	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "3",
	}
	doc.GenManTree(cmd, header, "/tmp")
}

func ExampleGenMan() {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "my test program",
	}
	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "3",
	}
	out := new(bytes.Buffer)
	doc.GenMan(cmd, header, out)
	fmt.Print(out.String())
}
