package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"

	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrationsFS embed.FS

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: migrate <file_path> up-to|down-to <version>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	direction := os.Args[2]
	versionStr := os.Args[3]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		log.Fatalf("Invalid version number: %v", err)
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %v", filePath)
	}
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		log.Fatalf("Failed to get migrations from embedded filesystem")
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations)
	if err != nil {
		log.Fatalf("Failed to create migration provider: %v", err)
	}

	ctx := context.Background()

	switch direction {
	case "up-to":
		_, err = provider.UpTo(ctx, int64(version))
	case "down-to":
		_, err = provider.DownTo(ctx, int64(version))
	default:
		log.Fatalf("Invalid direction: use 'up-to' or 'down-to'")
	}

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration successful!")
}
