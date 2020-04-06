package v2

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/RTradeLtd/swampi"
	"github.com/gin-gonic/gin"
)

// SwarmUpload is used to upload data to ethereum swarm
func (api *API) SwarmUpload(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("until further notice swarm based calls are not supported in production"))
		return
	}
	/* TODO(bonedaddy): dont auth gate for now so we can shill dank memes
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// temporary until we do database updates
	_ = username
	*/
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	// fileName := fileHandler.Filename
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	isTar := c.PostForm("is_tar")
	openFile, err := fileHandler.Open()
	if err != nil {
		Fail(c, err)
		return
	}
	fileBytes, err := ioutil.ReadAll(openFile)
	if err != nil {
		Fail(c, err)
		return
	}
	var ipfsHash string
	// store on ipfs if requested
	if c.PostForm("mirror_ipfs") == "true" {
		ipfsHash, err = api.ipfs.Add(bytes.NewReader(append(fileBytes[0:0:0], fileBytes...)))
		if err != nil {
			Fail(c, err)
			return
		}
	}
	// lazy so im hard coding this for now
	swamp := swampi.New("http://localhost:8500")
	resp, err := swamp.Send(swampi.SingleFileUpload, bytes.NewReader(append(fileBytes[0:0:0], fileBytes...)), map[string][]string{
		"content-type": {swampi.SingleFileUpload.ContentType(isTar == "true")},
	})
	if err != nil {
		api.LogError(c, err, err.Error())
		return
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Fail(c, err)
		return
	}
	resp.Body.Close()
	if string(contents) == "" {
		Fail(c, errors.New("bad contents returned from swarm api"))
		return
	}
	var response = make(gin.H)
	if ipfsHash != "" {
		response["response"] = map[string]string{
			"ipfs_hash":  ipfsHash,
			"swarm_hash": string(contents),
		}
	} else {
		response["response"] = map[string]string{
			"swarm_hash": string(contents),
		}
	}
	// TODO(bonedaddy): update database records
	Respond(c, http.StatusOK, response)
}
