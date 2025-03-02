package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"gosl/internal/bot"
	"gosl/internal/httpserver"
	"gosl/pkg/config"
	"gosl/pkg/embedfs"
	"gosl/pkg/logging"

	"github.com/pkg/errors"
)

var maint uint32 // atomic: 1 if in maintenance mode

// Initializes and runs the server
func run(ctx context.Context, w io.Writer, args map[string]string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := config.GetConfig(args)
	if err != nil {
		return errors.Wrap(err, "config.GetConfig")
	}

	// Return the version of the database required
	if args["dbver"] == "true" {
		fmt.Fprintf(w, "Database version: %s\n", config.DBName)
		return nil
	}

	// Setup the logfile
	var logfile *os.File = nil
	if config.LogOutput == "both" || config.LogOutput == "file" {
		logfile, err = logging.GetLogFile(config.LogDir)
		if err != nil {
			return errors.Wrap(err, "logging.GetLogFile")
		}
		defer logfile.Close()
	}

	// Setup the console writer
	var consoleWriter io.Writer
	if config.LogOutput == "both" || config.LogOutput == "console" {
		consoleWriter = w
	}

	// Setup the logger
	logger, err := logging.GetLogger(
		config.LogLevel,
		consoleWriter,
		logfile,
		config.LogDir,
	)
	if err != nil {
		return errors.Wrap(err, "logging.GetLogger")
	}

	// Setup the database connection
	logger.Debug().Msg("Config loaded and logger started")
	logger.Debug().Msg("Connecting to database")
	conn, err := setupDBConn(args, logger, config)
	if err != nil {
		return errors.Wrap(err, "setupDBConn")
	}
	defer conn.Close()

	// Setup embedded files
	logger.Debug().Msg("Getting embedded files")
	staticFS, err := embedfs.GetEmbeddedFS()
	if err != nil {
		return errors.Wrap(err, "embedfs.GetEmbeddedFS")
	}

	logger.Debug().Msg("Setting up HTTP server")
	httpServer := httpserver.NewServer(config, logger, conn, &staticFS, &maint)

	// Runs function for testing in dev if --tester flag true
	if args["tester"] == "true" {
		logger.Debug().Msg("Running tester function")
		test(config, logger, conn, httpServer)
		return nil
	}

	// Setups a channel to listen for os.Signal
	handleMaintSignals(ctx, conn, config, logger, &maint)

	// Initialize the discord bot
	discordBot, err := bot.NewBot(config.DiscordBotToken, logger, &staticFS)
	if err != nil {
		return errors.Wrap(err, "bot.NewBot")
	}

	// Runs the http server
	logger.Debug().Msg("Starting up the HTTP server")
	go func() {
		logger.Info().Str("address", httpServer.Addr).Msg("Listening for requests")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Error listening and serving")
		}
	}()

	// Runs the discord bot
	go func() {
		logger.Info().Msg("Starting discord bot")
		if err := discordBot.Start(); err != nil {
			logger.Error().Err(err).Msg("Error running bot")
		}
	}()

	// Handles graceful shutdown
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("Error shutting down http server")
		}
	}()
	go func() {
		defer wg.Done()
		<-ctx.Done()
		defer cancel()
		if err := discordBot.Stop(); err != nil {
			logger.Error().Err(err).Msg("Error shutting down discord bot")
		}
	}()
	wg.Wait()
	logger.Info().Msg("Shutting down")
	return nil

}
