package slapshotapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type endpointSteamID struct {
	steamID string
}

func getEndpointSteamID(steamID string) *endpointSteamID {
	return &endpointSteamID{
		steamID: steamID,
	}
}

func (ep *endpointSteamID) path() string {
	return fmt.Sprintf("/api/public/players/steam/%s", ep.steamID)
}

func (ep *endpointSteamID) method() string {
	return "GET"
}

type idresp struct {
	ID uint32 `json:"id"`
}

// Get the SlapID of the steam user
func GetSlapID(
	ctx context.Context,
	steamid string,
	cfg *SlapAPIConfig,
) (uint32, error) {
	endpoint := getEndpointSteamID(steamid)
	data, err := slapapiReq(ctx, endpoint, cfg)
	if err != nil {
		return 0, errors.Wrap(err, "slapapiReq")
	}
	resp := idresp{}
	json.Unmarshal(data, &resp)
	return resp.ID, nil
}
