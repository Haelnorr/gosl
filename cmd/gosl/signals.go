package main

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"gosl/pkg/config"
	"gosl/pkg/db"

	"github.com/rs/zerolog"
)

// Handle SIGUSR1 and SIGUSR2 syscalls to toggle maintenance mode
func handleMaintSignals(
	ctx context.Context,
	conn *db.SafeConn,
	config *config.Config,
	logger *zerolog.Logger,
	maint *uint32,
) {
	logger.Debug().Msg("Starting signal listener")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				logger.Debug().Msg("Shutting down signal listener")
				signal.Stop(ch)
				return
			case sig := <-ch:
				switch sig {
				case syscall.SIGUSR1:
					if atomic.LoadUint32(maint) != 1 {
						atomic.StoreUint32(maint, 1)
						logger.Info().Msg("Signal received: Starting maintenance")
						logger.Info().Msg("Attempting to acquire database lock")
						conn.Pause(config.DBLockTimeout * time.Second)
					}
				case syscall.SIGUSR2:
					if atomic.LoadUint32(maint) != 0 {
						logger.Info().Msg("Signal received: Maintenance over")
						logger.Info().Msg("Releasing database lock")
						conn.Resume()
						atomic.StoreUint32(maint, 0)
					}
				}
			}
		}
	}()
}
