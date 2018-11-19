package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/grpc/lens/request"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// submitIndexRequest is used to submit an object to be indexed by Lens
func (api *API) submitIndexRequest(c *gin.Context) {
	objectType, exists := c.GetPostForm("object_type")
	if !exists {
		FailWithMissingField(c, "object_type")
		return
	}
	objectIdentifier, exists := c.GetPostForm("object_identifier")
	if !exists {
		FailWithMissingField(c, "object_identifier")
		return
	}
	switch objectType {
	case "ipld":
		// validate the object identifier
		if _, err := gocid.Decode(objectIdentifier); err != nil {
			Fail(c, err)
			return
		}
	default:
		Fail(c, errors.New(eh.InvalidObjectTypeError))
		return
	}
	resp, err := api.lc.Index(context.Background(), &request.Index{
		DataType:         objectType,
		ObjectIdentifier: objectIdentifier,
	})
	if err != nil {
		api.LogError(err, eh.FailedToIndexError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{
		"response": gin.H{
			"lens_id":  resp.GetId(),
			"keywords": resp.GetKeywords(),
		},
	})
}

// submitSearchRequest is used to send a search request to lens
func (api *API) submitSearchRequest(c *gin.Context) {
	keywords, exists := c.GetPostFormArray("keywords")
	if !exists {
		FailWithMissingField(c, "keywords")
		return
	}
	var keywordsLower []string
	for _, word := range keywords {
		keywordsLower = append(keywordsLower, strings.ToLower(word))
	}
	resp, err := api.lc.Search(context.Background(), &request.Search{
		Keywords: keywordsLower,
	})
	if err != nil {
		fmt.Println(err)
		api.LogError(err, eh.FailedToSearchError)(c, http.StatusBadRequest)
		return
	}
	if len(resp.GetObjects()) == 0 {
		api.LogInfo(fmt.Sprintf("no search results found for keywords %s", keywordsLower))
		Respond(c, http.StatusBadRequest, gin.H{"response": "no results found"})
		return
	}
	Respond(c, http.StatusOK, gin.H{
		"response": resp.GetObjects(),
	})
}
