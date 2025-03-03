package tests

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

func findMigrations() (*fs.FS, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			migrationsdir := os.DirFS(filepath.Join(dir, "cmd", "migrate", "migrations"))
			return &migrationsdir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir { // Reached root
			return nil, errors.New("Unable to locate migrations directory")
		}
		dir = parent
	}
}

func findTestData() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			return filepath.Join(dir, "pkg", "tests", "testdata.sql"), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir { // Reached root
			return "", errors.New("Unable to locate test data")
		}
		dir = parent
	}
}

func migrateTestDB(wconn *sql.DB, version int64) error {
	migrations, err := findMigrations()
	if err != nil {
		return errors.Wrap(err, "findMigrations")
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, wconn, *migrations)
	if err != nil {
		return errors.Wrap(err, "goose.NewProvider")
	}
	ctx := context.Background()
	if _, err := provider.UpTo(ctx, version); err != nil {
		return errors.Wrap(err, "provider.UpTo")
	}
	return nil
}

func loadTestData(wconn *sql.DB) error {
	dataPath, err := findTestData()
	if err != nil {
		return errors.Wrap(err, "findSchema")
	}
	sqlBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return errors.Wrap(err, "os.ReadFile")
	}
	dataSQL := string(sqlBytes)

	_, err = wconn.Exec(dataSQL)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

// Returns two db connection handles. First is a readwrite connection, second
// is a read only connection
func SetupTestDB(version int64) (*sql.DB, *sql.DB, error) {
	opts := "_journal_mode=WAL&_synchronous=NORMAL&_txlock=IMMEDIATE"
	file := fmt.Sprintf("file::memory:?cache=shared&%s", opts)
	wconn, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, nil, errors.Wrap(err, "sql.Open")
	}

	err = migrateTestDB(wconn, version)
	if err != nil {
		return nil, nil, errors.Wrap(err, "migrateTestDB")
	}
	err = loadTestData(wconn)
	if err != nil {
		return nil, nil, errors.Wrap(err, "loadTestData")
	}

	opts = "_synchronous=NORMAL&mode=ro"
	file = fmt.Sprintf("file::memory:?cache=shared&%s", opts)
	rconn, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, nil, errors.Wrap(err, "sql.Open")
	}
	return wconn, rconn, nil
}
