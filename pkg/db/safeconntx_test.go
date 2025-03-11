package db

import (
	"context"
	"gosl/pkg/tests"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeConn(t *testing.T) {
	cfg, err := tests.TestConfig()
	require.NoError(t, err)
	logger := tests.NilLogger()
	ver, err := strconv.ParseInt(cfg.DBName, 10, 0)
	require.NoError(t, err)
	wconn, rconn, err := tests.SetupTestDB(ver)
	require.NoError(t, err)
	sconn := MakeSafe(wconn, rconn, logger)
	defer sconn.Close()

	t.Run("Global lock waits for read locks to finish", func(t *testing.T) {
		tx, err := sconn.Begin(t.Context())
		require.NoError(t, err)
		var requested sync.WaitGroup
		var engaged sync.WaitGroup
		requested.Add(1)
		engaged.Add(1)
		go func() {
			requested.Done()
			sconn.Pause(5 * time.Second)
			engaged.Done()
		}()
		requested.Wait()
		assert.Equal(t, uint32(0), sconn.globalLockStatus)
		assert.Equal(t, uint32(1), sconn.globalLockRequested)
		tx.Commit()
		engaged.Wait()
		assert.Equal(t, uint32(1), sconn.globalLockStatus)
		assert.Equal(t, uint32(0), sconn.globalLockRequested)
		sconn.Resume()
	})
	t.Run("Lock abandons after timeout", func(t *testing.T) {
		tx, err := sconn.Begin(t.Context())
		require.NoError(t, err)
		sconn.Pause(250 * time.Millisecond)
		assert.Equal(t, uint32(0), sconn.globalLockStatus)
		assert.Equal(t, uint32(0), sconn.globalLockRequested)
		tx.Commit()
	})
	t.Run("TX abandons after timeout", func(t *testing.T) {
		sconn.Pause(250 * time.Millisecond)
		ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
		defer cancel()
		_, err = sconn.Begin(ctx)
		require.Error(t, err)
		sconn.Resume()
	})
	t.Run("Pause blocks transactions and resume allows", func(t *testing.T) {
		tx, err := sconn.Begin(t.Context())
		require.NoError(t, err)
		var requested sync.WaitGroup
		var engaged sync.WaitGroup
		requested.Add(1)
		engaged.Add(1)
		go func() {
			requested.Done()
			sconn.Pause(5 * time.Second)
			engaged.Done()
		}()
		requested.Wait()
		assert.Equal(t, uint32(0), sconn.globalLockStatus)
		assert.Equal(t, uint32(1), sconn.globalLockRequested)
		ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
		defer cancel()
		_, err = sconn.Begin(ctx)
		require.Error(t, err)
		tx.Commit()
		engaged.Wait()
		_, err = sconn.Begin(ctx)
		require.Error(t, err)
		sconn.Resume()
		tx, err = sconn.Begin(t.Context())
		require.NoError(t, err)
		tx.Commit()
	})
}
func TestSafeTX(t *testing.T) {
	cfg, err := tests.TestConfig()
	require.NoError(t, err)
	logger := tests.NilLogger()
	ver, err := strconv.ParseInt(cfg.DBName, 10, 0)
	require.NoError(t, err)
	wconn, rconn, err := tests.SetupTestDB(ver)
	require.NoError(t, err)
	sconn := MakeSafe(wconn, rconn, logger)
	defer sconn.Close()

	t.Run("Commit releases lock", func(t *testing.T) {
		tx, err := sconn.Begin(t.Context())
		require.NoError(t, err)
		assert.Equal(t, uint32(1), sconn.readLockCount)
		tx.Commit()
		assert.Equal(t, uint32(0), sconn.readLockCount)
	})
	t.Run("Rollback releases lock", func(t *testing.T) {
		tx, err := sconn.Begin(t.Context())
		require.NoError(t, err)
		assert.Equal(t, uint32(1), sconn.readLockCount)
		tx.Rollback()
		assert.Equal(t, uint32(0), sconn.readLockCount)
	})
	t.Run("Multiple RTX can gain read lock", func(t *testing.T) {
		tx1, err := sconn.RBegin(t.Context())
		require.NoError(t, err)
		tx2, err := sconn.RBegin(t.Context())
		require.NoError(t, err)
		tx3, err := sconn.RBegin(t.Context())
		require.NoError(t, err)
		tx1.Commit()
		tx2.Commit()
		tx3.Commit()
	})
	t.Run("Lock acquiring times out after timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
		defer cancel()
		sconn.acquireGlobalLock()
		defer sconn.releaseGlobalLock()
		_, err := sconn.Begin(ctx)
		require.Error(t, err)
	})
	t.Run("Lock acquires if lock released", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 250*time.Millisecond)
		defer cancel()
		sconn.acquireGlobalLock()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			tx, err := sconn.Begin(ctx)
			require.NoError(t, err)
			tx.Commit()
			wg.Done()
		}()
		sconn.releaseGlobalLock()
		wg.Wait()
	})
}
