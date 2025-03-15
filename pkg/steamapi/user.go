package steamapi

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type playersresponse struct {
	Response struct {
		Players []User `json:"players"`
	} `json:"response"`
}

type User struct {
	SteamID                  string `json:"steamid"`
	CommunityVisibilityState uint16 `json:"communityvisibilitystate"`
	ProfileState             uint16 `json:"profilestate"`
	PersonaName              string `json:"personaname"`
	ProfileURL               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	AvatarMedium             string `json:"avatarmedium"`
	AvatarFull               string `json:"avatarfull"`
	AvatarHash               string `json:"avatarhash"`
	LastLogoff               uint64 `json:"lastlogoff"`
	PersonaState             uint16 `json:"personastate"`
	RealName                 string `json:"realname"`
	PrimaryClanID            string `json:"primaryclanid"`
	TimeCreated              uint64 `json:"timecreated"`
	PersonaStateFlags        uint64 `json:"personastateflags"`
	LocCountryCode           string `json:"loccountrycode"`
	LocStateCode             string `json:"locstatecode"`
	LocCityID                uint32 `json:"loccityid"`
}

// Get the player summary of the steam user from the Steam API
func GetUser(steamid string, apikey string) (*User, error) {
	endpoint := "ISteamUser/GetPlayerSummaries/v0002"
	opts := map[string]string{
		"steamids": steamid,
	}
	data, err := steamapiGet(endpoint, apikey, opts)
	if err != nil {
		return nil, errors.Wrap(err, "steamapiGet")
	}
	resp := playersresponse{}
	json.Unmarshal(data, &resp)
	if len(resp.Response.Players) == 0 {
		return nil, nil
	}
	user := resp.Response.Players[0]
	return &user, nil
}
