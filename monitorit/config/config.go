package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config stores all configuration required by monitorit.
type Config struct {
	DBPath              string
	AlertJobCron        string
	HTTPPort            string
	TwilioAccountSID    string
	TwilioAuthToken     string
	TwilioPhoneNumber   string
	AlertJobPhoneNumber string
}

// Read reads the configuration from the current environment.
// If a .env file is found in the current working directory,
// it will be read before reading environment variables.
func Read() (Config, error) {
	// Try reading .env
	switch err := godotenv.Load(".env"); {
	case errors.Is(err, fs.ErrNotExist):
		// File doesn't exist, do nothing and skip
	case err != nil:
		return Config{}, fmt.Errorf("failed to read .env file: %w", err)
	}

	var missing []string
	cfg := Config{
		DBPath:              requireEnv("DB_PATH", &missing),
		AlertJobCron:        requireEnv("ALERT_JOB_CRON", &missing),
		HTTPPort:            getEnv("HTTP_PORT", "8080"),
		TwilioAccountSID:    requireEnv("TWILIO_ACCOUNT_SID", &missing),
		TwilioAuthToken:     requireEnv("TWILIO_AUTH_TOKEN", &missing),
		TwilioPhoneNumber:   requireEnv("TWILIO_PHONE_NUMBER", &missing),
		AlertJobPhoneNumber: requireEnv("ALERT_JOB_PHONE_NUMBER", &missing),
	}
	if len(missing) > 0 {
		return cfg, fmt.Errorf("required env vars missing: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

// requireEnv gets and returns the env var for key.
// If the key does not exist, the it will be added to missing.
func requireEnv(key string, missing *[]string) string {
	v := os.Getenv(key)
	if v == "" {
		*missing = append(*missing, key)
	}
	return v
}

// getEnv gets and returns the env var for key.
// If the key does not exist, fallback is returned.
func getEnv(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
