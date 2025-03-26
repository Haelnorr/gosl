package httpserver

import (
	"net/http"

	"gosl/internal/handler"
	"gosl/pkg/config"
	"gosl/pkg/db"

	"github.com/rs/zerolog"
)

func addRoutes(
	mux *http.ServeMux,
	logger *zerolog.Logger,
	config *config.Config,
	conn *db.SafeConn,
	staticFS *http.FileSystem,
) {
	route := mux.Handle

	// Health check
	mux.HandleFunc("GET /healthz", func(http.ResponseWriter, *http.Request) {})

	// Static files
	route("GET /static/", http.StripPrefix("/static/", handler.HandleFS(staticFS)))

	// Index page and unhandled catchall (404)
	route("GET /", handler.Root())

	// Player Registration help page
	route("GET /registration-help", handler.RegistrationHelp())
}
