package db

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type SafeTX interface {
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) (*sql.Row, error)
	Commit() error
	Rollback() error
}

// Extends sql.Tx for use with SafeConn
type SafeWTX struct {
	tx    *sql.Tx
	sc    *SafeConn
	label string
}

type SafeRTX struct {
	tx    *sql.Tx
	sc    *SafeConn
	label string
}

func isWriteOperation(query string) bool {
	query = strings.TrimSpace(query)
	query = strings.ToUpper(query)
	writeOpsRegex := `^(INSERT|UPDATE|DELETE|REPLACE|MERGE|CREATE|DROP|ALTER|TRUNCATE)\s+`
	re := regexp.MustCompile(writeOpsRegex)
	return re.MatchString(query)
}

// Query the database inside the transaction
func (stx *SafeRTX) Query(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	if stx.tx == nil {
		return nil, errors.New("Cannot query without a transaction")
	}
	if isWriteOperation(query) {
		return nil, errors.New("Cannot query with a write operation")
	}
	rows, err := stx.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryContext")
	}
	return rows, nil
}

// Query the database inside the transaction
func (stx *SafeWTX) Query(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	if stx.tx == nil {
		return nil, errors.New("Cannot query without a transaction")
	}
	if isWriteOperation(query) {
		return nil, errors.New("Cannot query with a write operation")
	}
	rows, err := stx.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryContext")
	}
	return rows, nil
}

// Query a row from the database inside the transaction
func (stx *SafeRTX) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Row, error) {
	if stx.tx == nil {
		return nil, errors.New("Cannot query without a transaction")
	}
	if isWriteOperation(query) {
		return nil, errors.New("Cannot query with a write operation")
	}
	return stx.tx.QueryRowContext(ctx, query, args...), nil
}

// Query a row from the database inside the transaction
func (stx *SafeWTX) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Row, error) {
	if stx.tx == nil {
		return nil, errors.New("Cannot query without a transaction")
	}
	if isWriteOperation(query) {
		return nil, errors.New("Cannot query with a write operation")
	}
	return stx.tx.QueryRowContext(ctx, query, args...), nil
}

// Exec a statement on the database inside the transaction
func (stx *SafeWTX) Exec(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	if stx.tx == nil {
		return nil, errors.New("Cannot exec without a transaction")
	}
	res, err := stx.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tx.ExecContext")
	}
	return res, nil
}

// Commit the current transaction and release the read lock
func (stx *SafeRTX) Commit() error {
	if stx.tx == nil {
		return errors.New("Cannot commit without a transaction")
	}
	err := stx.tx.Commit()
	stx.tx = nil
	stx.sc.releaseReadLock(stx.label)
	return err
}

// Commit the current transaction and release the read lock
func (stx *SafeWTX) Commit() error {
	if stx.tx == nil {
		return errors.New("Cannot commit without a transaction")
	}
	err := stx.tx.Commit()
	stx.tx = nil
	stx.sc.releaseReadLock(stx.label)
	return err
}

// Abort the current transaction, releasing the read lock
func (stx *SafeRTX) Rollback() error {
	if stx.tx == nil {
		return errors.New("Cannot rollback without a transaction")
	}
	err := stx.tx.Rollback()
	stx.tx = nil
	stx.sc.releaseReadLock(stx.label)
	return err
}

// Abort the current transaction, releasing the read lock
func (stx *SafeWTX) Rollback() error {
	if stx.tx == nil {
		return errors.New("Cannot rollback without a transaction")
	}
	err := stx.tx.Rollback()
	stx.tx = nil
	stx.sc.releaseReadLock(stx.label)
	return err
}
