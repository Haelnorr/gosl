package db

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	_ "modernc.org/sqlite"
)

// Returns a database connection handle for the DB
func ConnectToDatabase(
	dbName string,
	logger *zerolog.Logger,
) (*SafeConn, error) {
	file := fmt.Sprintf("file:%s.db", dbName)
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open")
	}
	version, err := strconv.Atoi(dbName)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.Atoi")
	}
	err = checkDBVersion(db, version)
	if err != nil {
		return nil, errors.Wrap(err, "checkDBVersion")
	}
	conn := MakeSafe(db, logger)
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
