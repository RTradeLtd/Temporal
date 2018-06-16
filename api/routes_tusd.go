package api

import (
	"github.com/tus/tusd"
	"github.com/tus/tusd/filestore"
)

/*
Experimental TUS implementation
Initially it will be strictly to store files on disk but will expand to include IPFS as a backend
*/

func generateTUSHandler() (*tusd.Handler, error) {
	store := filestore.FileStore{
		Path: "./uploads",
	}

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:      "/home/solidity",
		StoreComposer: composer,
	})

	if err != nil {
		return nil, err
	}

	return handler, nil
}
