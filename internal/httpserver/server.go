package httpserver

import (
	"io/fs"
	"net"
	"net/http"
	"time"

	"gosl/internal/middleware"
	"gosl/pkg/config"
	"gosl/pkg/db"

	"github.com/rs/zerolog"
)

func NewServer(
	config *config.Config,
	logger *zerolog.Logger,
	conn *db.SafeConn,
	staticFS *fs.FS,
	maint *uint32,
) *http.Server {
	fs := http.FS(*staticFS)
	srv := createServer(config, logger, conn, &fs, maint)
	httpServer := &http.Server{
		Addr:              net.JoinHostPort(config.Host, config.Port),
		Handler:           srv,
		ReadHeaderTimeout: config.ReadHeaderTimeout * time.Second,
		WriteTimeout:      config.WriteTimeout * time.Second,
		IdleTimeout:       config.IdleTimeout * time.Second,
	}
	return httpServer
}

// Returns a new http.Handler with all the routes and middleware added
func createServer(
	config *config.Config,
	logger *zerolog.Logger,
	conn *db.SafeConn,
	staticFS *http.FileSystem,
	maint *uint32,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		logger,
		config,
		conn,
		staticFS,
	)
	var handler http.Handler = mux
	// Add middleware here, must be added in reverse order of execution
	// i.e. First in list will get executed last during the request handling
	handler = middleware.Logging(logger, handler)
	// handler = middleware.Authentication(logger, config, conn, handler, maint)

	// Gzip
	// handler = middleware.Gzip(handler, config.GZIP)

	// Start the timer for the request chain so logger can have accurate info
	handler = middleware.StartTimer(handler)
	return handler
}
