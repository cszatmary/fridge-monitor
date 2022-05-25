package main

import (
	"database/sql"
	"log"

	"github.com/cszatmary/fridge-monitor/monitorit/config"
	"github.com/cszatmary/fridge-monitor/monitorit/jobs"
	"github.com/cszatmary/fridge-monitor/monitorit/lib/sms"
	"github.com/cszatmary/fridge-monitor/monitorit/models"
	"github.com/cszatmary/fridge-monitor/monitorit/routes"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Setup the SQLite DB
	db, err := sql.Open("sqlite3", cfg.DBPath)
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

	// Initialize dependencies
	fm := models.NewFridgeManager(db)
	tm := models.NewTemperatureManager(db)
	smsClient := sms.NewClient(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioPhoneNumber)

	// Setup job runner
	s := jobs.Setup(jobs.SetupDependencies{
		AlertJobCron:        cfg.AlertJobCron,
		FridgeManager:       fm,
		TemperatureManager:  tm,
		SMSClient:           smsClient,
		AlertJobPhoneNumber: cfg.AlertJobPhoneNumber,
	})
	s.StartAsync()
	log.Print("Job runner started")

	// Setup HTTP server
	app := routes.SetupApp(routes.SetupDependencies{
		DB:                 db,
		FridgeManager:      fm,
		TemperatureManager: tm,
	})
	log.Fatal(app.Listen(":" + cfg.HTTPPort))
}
