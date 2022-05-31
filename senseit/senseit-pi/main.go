package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/d2r2/go-bsbmp"
	"github.com/d2r2/go-i2c"
	loggerpkg "github.com/d2r2/go-logger"
)

func main() {
	if err := execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func execute() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("sensor address and monitorit post URL required as arguments")
	}

	// Parse the address as an uint, usually in hex
	sensorAddress, err := strconv.ParseUint(os.Args[1], 0, 8)
	if err != nil {
		return fmt.Errorf("failed to parse sensor address as a uint: %w", err)
	}
	monitoritURL := os.Args[2]

	// Create new connection to i2c-bus on 1 line with the given address.
	i2c, err := i2c.NewI2C(uint8(sensorAddress), 1)
	if err != nil {
		return fmt.Errorf("failed to create i2c connection to sensor: %w", err)
	}
	defer i2c.Close()
	// Supress verbose output from i2c package.
	if err := loggerpkg.ChangePackageLogLevel("i2c", loggerpkg.InfoLevel); err != nil {
		return fmt.Errorf("failed to set log level for i2c package: %w", err)
	}

	sensor, err := bsbmp.NewBMP(bsbmp.BME280, i2c)
	if err != nil {
		return fmt.Errorf("failed to initialize sensor: %w", err)
	}
	// Suppress verbose output from the bsbmp package.
	if err := loggerpkg.ChangePackageLogLevel("bsbmp", loggerpkg.InfoLevel); err != nil {
		return fmt.Errorf("failed to set log level for bsbmp package: %w", err)
	}

	// Read the temperature and send it to monitorit
	temp, err := sensor.ReadTemperatureC(bsbmp.ACCURACY_STANDARD)
	if err != nil {
		return fmt.Errorf("failed to read temperature from sensor: %w", err)
	}

	body := struct {
		Value float32 `json:"value"`
	}{Value: temp}
	var bodyBuf bytes.Buffer
	if err := json.NewEncoder(&bodyBuf).Encode(body); err != nil {
		return fmt.Errorf("failed to encode request body as JSON: %w", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Post(monitoritURL, "application/json", &bodyBuf)
	if err != nil {
		return fmt.Errorf("failed to send POST request to monitorit: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("received status %d from monitorit", resp.StatusCode)
	}
	return nil
}
