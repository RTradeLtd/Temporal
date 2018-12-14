// Package importer implements utilities used to create IPFS DAGs from files
// and readers.
package importer

import (
	bal "gx/ipfs/QmdYvDbHp7qAhZ7GsCj6e1cMo55ND6y2mjWVzwdvcv4f12/go-unixfs/importer/balanced"
	h "gx/ipfs/QmdYvDbHp7qAhZ7GsCj6e1cMo55ND6y2mjWVzwdvcv4f12/go-unixfs/importer/helpers"
	trickle "gx/ipfs/QmdYvDbHp7qAhZ7GsCj6e1cMo55ND6y2mjWVzwdvcv4f12/go-unixfs/importer/trickle"

	chunker "gx/ipfs/QmR4QQVkBZsZENRjYFVi8dEtPL3daZRNKk24m4r6WKJHNm/go-ipfs-chunker"
	ipld "gx/ipfs/QmcKKBwfz6FyQdHR2jsXrrF6XeSBXYL86anmWNewpFpoF5/go-ipld-format"
)

// BuildDagFromReader creates a DAG given a DAGService and a Splitter
// implementation (Splitters are io.Readers), using a Balanced layout.
func BuildDagFromReader(ds ipld.DAGService, spl chunker.Splitter) (ipld.Node, error) {
	dbp := h.DagBuilderParams{
		Dagserv:  ds,
		Maxlinks: h.DefaultLinksPerBlock,
	}

	return bal.Layout(dbp.New(spl))
}

// BuildTrickleDagFromReader creates a DAG given a DAGService and a Splitter
// implementation (Splitters are io.Readers), using a Trickle Layout.
func BuildTrickleDagFromReader(ds ipld.DAGService, spl chunker.Splitter) (ipld.Node, error) {
	dbp := h.DagBuilderParams{
		Dagserv:  ds,
		Maxlinks: h.DefaultLinksPerBlock,
	}

	return trickle.Layout(dbp.New(spl))
}
