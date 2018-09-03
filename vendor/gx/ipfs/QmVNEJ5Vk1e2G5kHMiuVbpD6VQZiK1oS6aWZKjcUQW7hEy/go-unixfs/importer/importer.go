// Package importer implements utilities used to create IPFS DAGs from files
// and readers.
package importer

import (
	bal "gx/ipfs/QmVNEJ5Vk1e2G5kHMiuVbpD6VQZiK1oS6aWZKjcUQW7hEy/go-unixfs/importer/balanced"
	h "gx/ipfs/QmVNEJ5Vk1e2G5kHMiuVbpD6VQZiK1oS6aWZKjcUQW7hEy/go-unixfs/importer/helpers"
	trickle "gx/ipfs/QmVNEJ5Vk1e2G5kHMiuVbpD6VQZiK1oS6aWZKjcUQW7hEy/go-unixfs/importer/trickle"

	ipld "gx/ipfs/QmX5CsuHyVZeTLxgRSYkgLSDQKb9UjE8xnhQzCEJWWWFsC/go-ipld-format"
	chunker "gx/ipfs/QmXzBbJo2sLf3uwjNTeoWYiJV7CjAhkiA4twtLvwJSSNdK/go-ipfs-chunker"
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
