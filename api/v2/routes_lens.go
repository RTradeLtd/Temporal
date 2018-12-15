package v2

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
	forms := api.extractPostForms(c, "object_type", "object_identifier")
	if len(forms) == 0 {
		return
	}
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
	resp, err := api.lens.Index(context.Background(), &request.Index{
		DataType:         forms["object_type"],
		ObjectIdentifier: forms["object_identifier"],
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
	resp, err := api.lens.Search(context.Background(), &request.Search{
		Keywords: keywordsLower,
	})
	if err != nil {
		fmt.Println(err)
		api.LogError(err, eh.FailedToSearchError)(c, http.StatusBadRequest)
		return
	}
	if len(resp.GetObjects()) == 0 {
		api.l.Info(fmt.Sprintf("no search results found for keywords %s", keywordsLower))
		Respond(c, http.StatusBadRequest, gin.H{"response": "no results found"})
		return
	}
	Respond(c, http.StatusOK, gin.H{
		"response": resp.GetObjects(),
	})
}
