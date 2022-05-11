package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/cszatmary/fridge-monitor/monitorit/routes"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("DB_PATH env var is not set")
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to create db instance: %v", err)
	}
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://database/migrations", "sqlite3", driver)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Apply all migrations
	switch err := m.Up(); err {
	// Up will error if migrations were already applied, we want it to no-op
	case nil, migrate.ErrNoChange:
	default:
		log.Fatalf("Failed to apply db migrations: %v", err)
	}

	app := routes.SetupApp(db)
	log.Fatal(app.Listen(":8080"))
}
