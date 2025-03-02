package tests

import (
	"gosl/pkg/config"
	"os"

	"github.com/pkg/errors"
)

func TestConfig() (*config.Config, error) {
	os.Setenv("SECRET_KEY", ".")
	os.Setenv("DISCORD_BOT_TOKEN", ".")
	cfg, err := config.GetConfig(map[string]string{})
	if err != nil {
		return nil, errors.Wrap(err, "config.GetConfig")
	}
	return cfg, nil
}
