package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tus/tusd"
	"github.com/tus/tusd/filestore"
)

/*
Experimental TUS implementation
Initially it will be strictly to store files on disk but will expand to include IPFS as a backend
*/

func TUSFileStore(c *gin.Context) {

}

func generateTUSHandler() (http.Handler, error) {
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

	return handler.Handler, nil
}
