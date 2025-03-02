package main

import (
	"net/http"

	"gosl/pkg/config"
	"gosl/pkg/db"

	"github.com/rs/zerolog"
)

// This function will only be called if the --test commandline flag is set.
// After the function finishes the application will close.
// Running command `make tester` will run the test using port 3232 to avoid
// conflicts on the default 3333. Useful for testing things out during dev.
// If you add code here, remember to run:
// `git update-index --assume-unchanged tester.go` to avoid tracking changes
func test(
	config *config.Config,
	logger *zerolog.Logger,
	conn *db.SafeConn,
	srv *http.Server,
) {
}
