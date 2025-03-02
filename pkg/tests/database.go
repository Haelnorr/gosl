package tests

import (
	"context"
	"database/sql"
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

func SetupTestDB(version int64) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open")
	}

	migrations, err := findMigrations()
	if err != nil {
		return nil, errors.Wrap(err, "findMigrations")
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, conn, *migrations)
	if err != nil {
		return nil, errors.Wrap(err, "goose.NewProvider")
	}
	ctx := context.Background()
	if _, err := provider.UpTo(ctx, version); err != nil {
		return nil, errors.Wrap(err, "provider.UpTo")
	}

	// Load the test data
	dataPath, err := findTestData()
	if err != nil {
		return nil, errors.Wrap(err, "findSchema")
	}
	sqlBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, errors.Wrap(err, "os.ReadFile")
	}
	dataSQL := string(sqlBytes)

	_, err = conn.Exec(dataSQL)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Exec")
	}
	return conn, nil
}
