package slapshotapi

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type idresp struct {
	ID uint32 `json:"id"`
}

// Get the SlapID of the steam user
func GetSlapID(steamid string, apikey string, env string) (uint32, error) {
	endpoint := "api/public/players/steam/%s"
	data, err := slapapiGet(fmt.Sprintf(endpoint, steamid), env, apikey)
	if err != nil {
		return 0, errors.Wrap(err, "slapapiGet")
	}
	resp := idresp{}
	json.Unmarshal(data, &resp)
	return resp.ID, nil
}
