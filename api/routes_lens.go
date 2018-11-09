package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Temporal/eh"
	pb "github.com/RTradeLtd/grpc/lens/request"
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
	req := &pb.IndexRequest{
		DataType:         objectType,
		ObjectIdentifier: objectIdentifier,
	}
	resp, err := api.lc.SubmitIndexRequest(context.Background(), req)
	if err != nil {
		api.LogError(err, eh.FailedToIndexError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"lens_id": resp.LensIdentifier, "keywords": resp.Keywords}})
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
	req := &pb.SearchRequest{
		Keywords: keywordsLower,
	}
	resp, err := api.lc.SubmitSearchRequest(context.Background(), req)
	if err != nil {
		api.LogError(err, eh.FailedToSearchError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": resp.Names})
}
