package v2

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/grpc/lens/request"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// submitIndexRequest is used to submit an object to be indexed by Lens
func (api *API) submitIndexRequest(c *gin.Context) {
	// extract post forms
	forms, missingField := api.extractPostForms(c, "object_type", "object_identifier")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// ensure the type being requested is supported
	switch forms["object_type"] {
	case "ipld":
		// validate the object identifier
		if _, err := gocid.Decode(forms["object_identifier"]); err != nil {
			Fail(c, err)
			return
		}
	default:
		Fail(c, errors.New(eh.InvalidObjectTypeError))
		return
	}
	// send the index request
	resp, err := api.lens.Index(context.Background(), &request.Index{
		Type:       forms["object_type"],
		Identifier: forms["object_identifier"],
		// allow for optional reindexing
		Reindex: c.PostForm("reindex") == "yes",
	})
	if err != nil {
		api.LogError(c, err, eh.FailedToIndexError)(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"lens_id":  resp.GetId(),
			"keywords": resp.GetKeywords(),
		},
	})
}

// submitSearchRequest is used to send a search request to lens
func (api *API) submitSearchRequest(c *gin.Context) {
	// extract key words to search for
	keywords, exists := c.GetPostFormArray("keywords")
	if !exists {
		FailWithMissingField(c, "keywords")
		return
	}
	// for all keywords to lower-case
	var keywordsLower []string
	for _, word := range keywords {
		keywordsLower = append(keywordsLower, strings.ToLower(word))
	}
	// send search query
	resp, err := api.lens.Search(context.Background(), &request.Search{
		Keywords: keywordsLower,
	})
	if err != nil {
		api.LogError(c, err, eh.FailedToSearchError)(http.StatusBadRequest)
		return
	}
	// check to ensure some objects were found, otherwise log a warning
	if len(resp.GetObjects()) == 0 {
		api.l.Warnf("no search results found for keywords %s", keywordsLower)
		Respond(c, http.StatusBadRequest, gin.H{"response": "no results found"})
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{
		"response": resp.GetObjects(),
	})
}
