package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type SafeConn struct {
	wconn               *sql.DB
	rconn               *sql.DB
	readLockCount       uint32
	globalLockStatus    uint32
	globalLockRequested uint32
	logger              *zerolog.Logger
}

// Make the provided db handle safe and attach a logger to it
func MakeSafe(wconn *sql.DB, rconn *sql.DB, logger *zerolog.Logger) *SafeConn {
	return &SafeConn{wconn: wconn, rconn: rconn, logger: logger}
}

// Attempts to acquire a global lock on the database connection
func (conn *SafeConn) acquireGlobalLock() bool {
	if conn.readLockCount > 0 || conn.globalLockStatus == 1 {
		return false
	}
	conn.globalLockStatus = 1
	conn.logger.Debug().Uint32("global_lock_status", conn.globalLockStatus).
		Msg("Global lock acquired")
	return true
}

// Releases a global lock on the database connection
func (conn *SafeConn) releaseGlobalLock() {
	conn.globalLockStatus = 0
	conn.logger.Debug().Uint32("global_lock_status", conn.globalLockStatus).
		Msg("Global lock released")
}

// Acquire a read lock on the connection. Multiple read locks can be acquired
// at the same time
func (conn *SafeConn) acquireReadLock(label string) bool {
	if conn.globalLockStatus == 1 || conn.globalLockRequested == 1 {
		return false
	}
	conn.readLockCount += 1
	conn.logger.Debug().Uint32("read_lock_count", conn.readLockCount).
		Str("label", label).Msg("Read lock acquired")
	return true
}

// Release a read lock. Decrements read lock count by 1
func (conn *SafeConn) releaseReadLock(label string) {
	conn.readLockCount -= 1
	conn.logger.Debug().Uint32("read_lock_count", conn.readLockCount).
		Str("label", label).Msg("Read lock released")
}

// Starts a new transaction based on the current context. Will cancel if
// the context is closed/cancelled/done
func (conn *SafeConn) Begin(ctx context.Context, label string) (*SafeWTX, error) {
	lockAcquired := make(chan struct{})
	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-lockCtx.Done():
			return
		default:
			if conn.acquireReadLock(label) {
				close(lockAcquired)
			}
		}
	}()

	select {
	case <-lockAcquired:
		tx, err := conn.wconn.BeginTx(ctx, nil)
		if err != nil {
			conn.releaseReadLock(label)
			return nil, err
		}
		return &SafeWTX{tx: tx, sc: conn, label: label}, nil
	case <-ctx.Done():
		cancel()
		return nil, errors.New("Transaction time out due to database lock")
	}
}

// Starts a new READONLY transaction based on the current context. Will cancel if
// the context is closed/cancelled/done
func (conn *SafeConn) RBegin(ctx context.Context, label string) (*SafeRTX, error) {
	lockAcquired := make(chan struct{})
	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-lockCtx.Done():
			return
		default:
			if conn.acquireReadLock(label) {
				close(lockAcquired)
			}
		}
	}()

	select {
	case <-lockAcquired:
		tx, err := conn.rconn.BeginTx(ctx, nil)
		if err != nil {
			conn.releaseReadLock(label)
			return nil, err
		}
		return &SafeRTX{tx: tx, sc: conn, label: label}, nil
	case <-ctx.Done():
		cancel()
		return nil, errors.New("Transaction time out due to database lock")
	}
}

// Acquire a global lock, preventing all transactions
func (conn *SafeConn) Pause(timeoutAfter time.Duration) {
	conn.logger.Info().Msg("Attempting to acquire global database lock")
	conn.globalLockRequested = 1
	defer func() { conn.globalLockRequested = 0 }()
	timeout := time.After(timeoutAfter)
	attempt := 0
	for {
		if conn.acquireGlobalLock() {
			conn.logger.Info().Msg("Global database lock acquired")
			return
		}
		select {
		case <-timeout:
			conn.logger.Info().Msg("Timeout: Global database lock abandoned")
			return
		case <-time.After(100 * time.Millisecond):
			attempt++
		}
	}
}

// Release the global lock
func (conn *SafeConn) Resume() {
	conn.releaseGlobalLock()
	conn.logger.Info().Msg("Global database lock released")
}

// Close the database connection
func (conn *SafeConn) Close() error {
	conn.logger.Debug().Msg("Acquiring global lock for connection close")
	conn.acquireGlobalLock()
	defer conn.releaseGlobalLock()
	conn.logger.Debug().Msg("Closing database connection")
	return conn.wconn.Close()
}
