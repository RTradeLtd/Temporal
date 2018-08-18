package rtfs

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

type DHTFindProvsResponse struct {
	ID        string `json:"id,omitempty"`
	Type      int    `json:"type,omitempty"`
	Responses [][]struct {
		ID    string   `json:"id,omitempty"`
		Addrs []string `json:"addrs,omitempty"`
	} `json:"responses,omitempty"`
	Extra string `json:"extra,omitempty"`
}
