package slapshotapi

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type SlapAPIConfig struct {
	env    string         // Environment; 'api' or 'staging'
	key    string         // API Key
	client *slapAPIClient // Rate limited http client
}

type endpoint interface {
	path() string
	method() string
}

func NewSlapAPIConfig(env, key string) (*SlapAPIConfig, error) {
	if env != "api" && env != "staging" {
		return nil, errors.New("Invalid Env specified, must be 'api' or 'staging'")
	}
	return &SlapAPIConfig{
		env:    env,
		key:    key,
		client: newRateLimitedClient(),
	}, nil
}

func slapapiReq(
	ctx context.Context,
	ep endpoint,
	cfg *SlapAPIConfig,
) ([]byte, error) {
	baseurl := fmt.Sprintf("https://%s.slapshot.gg%s", cfg.env, ep.path())
	req, err := http.NewRequest(ep.method(), baseurl, nil)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest")
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.key))
	res, err := cfg.client.Do(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "http.DefaultClient.Do")
	}
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Error making request: %v", res.StatusCode))
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "io.ReadAll")
	}
	return body, nil
}
