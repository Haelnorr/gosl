package db

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

// Extends sql.Tx for use with SafeConn
type SafeTX struct {
	tx *sql.Tx
	sc *SafeConn
}

// Query the database inside the transaction
func (stx *SafeTX) Query(
	ctx context.Context,
	query string,
	args ...interface{},
) (*sql.Rows, error) {
	if stx.tx == nil {
		return nil, errors.New("Cannot query without a transaction")
	}
	return stx.tx.QueryContext(ctx, query, args...)
}

// Exec a statement on the database inside the transaction
func (stx *SafeTX) Exec(
	ctx context.Context,
	query string,
	args ...interface{},
) (sql.Result, error) {
	if stx.tx == nil {
		return nil, errors.New("Cannot exec without a transaction")
	}
	return stx.tx.ExecContext(ctx, query, args...)
}

// Commit the current transaction and release the read lock
func (stx *SafeTX) Commit() error {
	if stx.tx == nil {
		return errors.New("Cannot commit without a transaction")
	}
	err := stx.tx.Commit()
	stx.tx = nil

	stx.sc.releaseReadLock()
	return err
}

// Abort the current transaction, releasing the read lock
func (stx *SafeTX) Rollback() error {
	if stx.tx == nil {
		return errors.New("Cannot rollback without a transaction")
	}
	err := stx.tx.Rollback()
	stx.tx = nil
	stx.sc.releaseReadLock()
	return err
}
