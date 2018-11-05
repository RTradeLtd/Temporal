package rtfs

// DefaultFSKeystorePath is the default location of a file-system based keystore
var DefaultFSKeystorePath = "/ipfs/keystore"

/*
{
    "ID": "<string>"
    "Type": "<int>"
    "Responses": [
        {
            "ID": "<string>"
            "Addrs": [
                "<object>"
            ]
        }
    ]
    "Extra": "<string>"
}

*/

// DHTFindProvsResponse is a response from the dht/findprovs api call
type DHTFindProvsResponse struct {
	ID        string `json:"id,omitempty"`
	Type      int    `json:"type,omitempty"`
	Responses [][]struct {
		ID    string   `json:"id,omitempty"`
		Addrs []string `json:"addrs,omitempty"`
	} `json:"responses,omitempty"`
	Extra string `json:"extra,omitempty"`
}
