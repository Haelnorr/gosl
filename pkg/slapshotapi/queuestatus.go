package slapshotapi

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type matchmakingresp struct {
	Playlists PubsQueue `json:"playlists"`
}

type PubsQueue struct {
	InQueue uint16 `json:"in_queue"`
	InMatch uint16 `json:"in_match"`
}

// Get the SlapID of the steam user
func GetQueueStatus(apikey string, env string) (*PubsQueue, error) {
	endpoint := "api/public/matchmaking?regions=oce-east"
	data, err := slapapiGet(endpoint, env, apikey)
	if err != nil {
		return nil, errors.Wrap(err, "slapapiGet")
	}
	resp := matchmakingresp{}
	json.Unmarshal(data, &resp)
	return &resp.Playlists, nil
}
