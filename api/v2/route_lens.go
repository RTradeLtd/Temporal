package v2

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Temporal/eh"
	pb "github.com/RTradeLtd/grpc/lensv2"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

func (api *API) submitIndexRequest(c *gin.Context) {
	// extract post forms
	forms, missingField := api.extractPostForms(c, "object_type", "object_identifier")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	var indexType pb.IndexReq_Type
	// ensure the type being requested is supported
	switch forms["object_type"] {
	case "ipld":
		// validate the object identifier
		if _, err := gocid.Decode(forms["object_identifier"]); err != nil {
			Fail(c, err)
			return
		}
		indexType = pb.IndexReq_IPLD
	default:
		Fail(c, errors.New(eh.InvalidObjectTypeError))
		return
	}

	resp, err := api.lens.Index(context.Background(), &pb.IndexReq{
		Type: indexType,
		Hash: forms["object_identifier"],
		Options: &pb.IndexReq_Options{
			Reindex: c.PostForm("reindex") == "yes",
		},
	})
	if err != nil {
		api.LogError(c, err, eh.FailedToIndexError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

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
	resp, err := api.lens.Search(context.Background(), &pb.SearchReq{
		Options: &pb.SearchReq_Options{
			Tags: keywordsLower,
		},
	})
	if err != nil {
		api.LogError(c, err, eh.FailedToSearchError)(http.StatusBadRequest)
		return
	}
	// check to ensure some objects were found, otherwise log a warning
	if len(resp.Results) == 0 {
		api.l.Warnf("no search results found for keywords %s", keywordsLower)
		Respond(c, http.StatusBadRequest, gin.H{"response": "no results found"})
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": resp})
}
