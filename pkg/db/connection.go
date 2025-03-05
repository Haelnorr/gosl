package db

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	_ "github.com/mattn/go-sqlite3"
)

// Returns a database connection handle for the DB
func ConnectToDatabase(
	dbName string,
	logger *zerolog.Logger,
) (*SafeConn, error) {
	opts := "_journal_mode=WAL&_synchronous=NORMAL&_txlock=IMMEDIATE"
	file := fmt.Sprintf("file:%s.db?%s", dbName, opts)
	wconn, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open (rw)")
	}
	wconn.SetMaxOpenConns(1)
	opts = "_synchronous=NORMAL&mode=ro"
	file = fmt.Sprintf("file:%s.db?%s", dbName, opts)

	rconn, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open (ro)")
	}

	version, err := strconv.Atoi(dbName)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.Atoi")
	}
	err = checkDBVersion(rconn, version)
	if err != nil {
		return nil, errors.Wrap(err, "checkDBVersion")
	}
	conn := MakeSafe(wconn, rconn, logger)
	return conn, nil
}

// Check the database version
func checkDBVersion(db *sql.DB, expectVer int) error {
	query := `SELECT version_id FROM goose_db_version WHERE is_applied = 1
    ORDER BY version_id DESC LIMIT 1`
	rows, err := db.Query(query)
	if err != nil {
		return errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	if rows.Next() {
		var version int
		err = rows.Scan(&version)
		if err != nil {
			return errors.Wrap(err, "rows.Scan")
		}
		if version != expectVer {
			return errors.New("Version mismatch")
		}
	} else {
		return errors.New("No version found")
	}
	return nil
}
