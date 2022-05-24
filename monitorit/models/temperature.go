package models

import (
	"context"
	"database/sql"

	"github.com/cszatmary/fridge-monitor/monitorit/lib/apierror"
)

type Temperature struct {
	ID        int64
	Value     float64
	FridgeID  int64
	CreatedAt Time
}

const (
	StatusNormal  = "normal"
	StatusTooLow  = "tooLow"
	StatusTooHigh = "tooHigh"
)

func (t Temperature) Status(minTemp, maxTemp float64) string {
	switch {
	case t.Value < minTemp:
		return StatusTooLow
	case t.Value > maxTemp:
		return StatusTooHigh
	default:
		return StatusNormal
	}
}

type TemperatureManager struct {
	db *sql.DB
}

func NewTemperatureManager(db *sql.DB) *TemperatureManager {
	return &TemperatureManager{db}
}

func (tm *TemperatureManager) FindMostRecentByFridgeID(ctx context.Context, fridgeID int64, limit int) ([]Temperature, error) {
	const op = apierror.Op("models.TemperatureManager.FindMostRecentByFridgeID")
	rows, err := resolveRunner(ctx, tm.db).
		QueryContext(
			ctx,
			`SELECT id, value, fridge_id, created_at FROM temperatures WHERE fridge_id = ? ORDER BY created_at DESC LIMIT ?`,
			fridgeID,
			limit,
		)
	if err != nil {
		return nil, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"failed to retrieve temperatures",
			op,
		)
	}

	var temperatures []Temperature
	for rows.Next() {
		var t Temperature
		err := rows.Scan(
			&t.ID,
			&t.Value,
			&t.FridgeID,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, apierror.Wrap(
				err,
				apierror.CodeDatabase,
				"failed to scan temperature row",
				op,
			)
		}
		temperatures = append(temperatures, t)
	}
	if err := rows.Err(); err != nil {
		return nil, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"error occurred while iterating over temperature rows",
			op,
		)
	}
	return temperatures, nil
}

func (tm *TemperatureManager) InsertOne(ctx context.Context, fridgeID int64, value float64) (Temperature, error) {
	const op = apierror.Op("models.TemperatureManager.InsertOne")
	var newTemp Temperature
	err := requireTxn(ctx).
		QueryRowContext(
			ctx,
			`INSERT INTO temperatures(value, fridge_id) VALUES(?, ?) RETURNING id, value, fridge_id, created_at`,
			value,
			fridgeID,
		).
		Scan(
			&newTemp.ID,
			&newTemp.Value,
			&newTemp.FridgeID,
			&newTemp.CreatedAt,
		)
	if err != nil {
		return newTemp, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"failed to insert temperature row",
			op,
		)
	}
	return newTemp, nil
}
