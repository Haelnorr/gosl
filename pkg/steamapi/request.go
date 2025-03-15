package steamapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func steamapiGet(
	endpoint string,
	key string,
	options map[string]string,
) ([]byte, error) {
	base := "https://api.steampowered.com/%s/?key=%s%s"
	opts := ""
	for k, v := range options {
		opts = fmt.Sprintf("%s&%s=%s", opts, k, v)
	}
	url := fmt.Sprintf(base, endpoint, key, opts)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequest")
	}
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
