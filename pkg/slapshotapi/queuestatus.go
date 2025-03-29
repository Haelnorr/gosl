package slapshotapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type endpointMatchmaking struct {
	regions []string
}

func getEndpointMatchmaking(regions []string) *endpointMatchmaking {
	return &endpointMatchmaking{
		regions: regions,
	}
}

func (ep *endpointMatchmaking) path() string {
	path := "/api/public/matchmaking%s"
	filters := ""
	if len(ep.regions) > 0 {
		filters = "?regions="
		for i, region := range ep.regions {
			filters = filters + region
			if i+1 != len(ep.regions) {
				filters = filters + ","
			}
		}
	}
	return fmt.Sprintf(path, filters)
}

func (ep *endpointMatchmaking) method() string {
	return "GET"
}

type matchmakingresp struct {
	Playlists PubsQueue `json:"playlists"`
}

type PubsQueue struct {
	InQueue uint16 `json:"in_queue"`
	InMatch uint16 `json:"in_match"`
}

// Get the SlapID of the steam user
func GetQueueStatus(
	ctx context.Context,
	regions []string,
	cfg *SlapAPIConfig,
) (*PubsQueue, error) {
	endpoint := getEndpointMatchmaking(regions)
	data, err := slapapiReq(ctx, endpoint, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "slapapiReq")
	}
	resp := matchmakingresp{}
	json.Unmarshal(data, &resp)
	return &resp.Playlists, nil
}
