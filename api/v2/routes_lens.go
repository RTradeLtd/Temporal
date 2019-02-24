package v2

import (
	"context"
	"errors"
	"net/http"

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
			Reindex: c.PostForm("reindex") == "true",
		},
	})
	if err != nil {
		api.LogError(c, err, eh.FailedToIndexError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": resp.GetDoc()})
}

func (api *API) submitSearchRequest(c *gin.Context) {
	// query we will use to perform as the main search
	query, exists := c.GetPostForm("query")
	if !exists {
		FailWithMissingField(c, "query")
		return
	}

	tags, _ := c.GetPostFormArray("tags")
	categories, _ := c.GetPostFormArray("categories")
	mimeTypes, _ := c.GetPostFormArray("mime_types")
	hashes, _ := c.GetPostFormArray("hashes")
	required, _ := c.GetPostFormArray("required")

	resp, err := api.lens.Search(context.Background(), &pb.SearchReq{
		Query: query,
		Options: &pb.SearchReq_Options{
			Tags:       tags,
			Categories: categories,
			MimeTypes:  mimeTypes,
			Hashes:     hashes,
			Required:   required,
		},
	})

	if err != nil {
		api.LogError(c, err, eh.FailedToSearchError)(http.StatusBadRequest)
		return
	}
	// check to ensure some objects were found, otherwise log a warning
	if len(resp.Results) == 0 {
		Respond(c, http.StatusBadRequest, gin.H{"response": "no results found"})
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": resp})
}
