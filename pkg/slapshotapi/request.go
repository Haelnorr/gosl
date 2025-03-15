package slapshotapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func slapapiGet(
	endpoint string,
	env string,
	key string,
) ([]byte, error) {
	if env != "api" && env != "staging" {
		return nil, errors.New("Invalid Env specified, must be 'api' or 'staging'")
	}
	baseurl := fmt.Sprintf("https://%s.slapshot.gg/%s", env, endpoint)
	req, err := http.NewRequest("GET", baseurl, nil)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest")
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", key))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http.DefaultClient.Do")
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "io.ReadAll")
	}
	return body, nil
}
