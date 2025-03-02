package main

import (
	"strconv"

	"gosl/pkg/config"
	"gosl/pkg/db"
	"gosl/pkg/tests"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func setupDBConn(
	args map[string]string,
	logger *zerolog.Logger,
	config *config.Config,
) (*db.SafeConn, error) {
	if args["test"] == "true" {
		logger.Debug().Msg("Server in test mode, using test database")
		ver, err := strconv.ParseInt(config.DBName, 10, 0)
		if err != nil {
			return nil, errors.Wrap(err, "strconv.ParseInt")
		}
		testconn, err := tests.SetupTestDB(ver)
		if err != nil {
			return nil, errors.Wrap(err, "tests.SetupTestDB")
		}
		conn := db.MakeSafe(testconn, logger)
		return conn, nil
	} else {
		conn, err := db.ConnectToDatabase(config.DBName, logger)
		if err != nil {
			return nil, errors.Wrap(err, "db.ConnectToDatabase")
		}
		return conn, nil
	}
}
