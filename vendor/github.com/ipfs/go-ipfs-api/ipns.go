package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
)

type PublishResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Publish updates a mutable name to point to a given value
func (s *Shell) Publish(node string, value string) (*PublishResponse, error) {
	var pubResp PublishResponse
	args := []string{value}
	if node != "" {
		args = []string{node, value}
	}

	resp, err := s.newRequest(context.TODO(), "name/publish", args...).Send(s.httpcli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Output)
	json.Unmarshal(buf.Bytes(), &pubResp)
	return &pubResp, nil
}

// PublishWithResponse updates a mutable name to point to a given value
func (s *Shell) PublishWithResponse(contentHash string, lifetime string, ttl string, key string, resolve bool) (*PublishResponse, error) {
	var pubResp PublishResponse
	var resolveString string
	if contentHash != "" || lifetime != "" || ttl != "" || key != "" {
		return nil, errors.New("empty strings provided")
	}
	if resolve {
		resolveString = "true"
	} else {
		resolveString = "false"
	}
	args := []string{contentHash, resolveString, lifetime, ttl, key}
	resp, err := s.newRequest(context.TODO(), "name/publish", args...).Send(s.httpcli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Output)
	json.Unmarshal(buf.Bytes(), &pubResp)
	return &pubResp, nil
}

// Resolve gets resolves the string provided to an /ipfs/[hash]. If asked to
// resolve an empty string, resolve instead resolves the node's own /ipns value.
func (s *Shell) Resolve(id string) (string, error) {
	var resp *Response
	var err error
	if id != "" {
		resp, err = s.newRequest(context.Background(), "name/resolve", id).Send(s.httpcli)
	} else {
		resp, err = s.newRequest(context.Background(), "name/resolve").Send(s.httpcli)
	}
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	var out struct{ Path string }
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Path, nil
}
