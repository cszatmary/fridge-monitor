package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/cszatmary/fridge-monitor/monitorit/lib/apierror"
)

type Fridge struct {
	ID            int64
	Name          string
	Description   string
	MinTemp       float64
	MaxTemp       float64
	AlertsEnabled bool
}

type FridgeManager struct {
	db *sql.DB
}

func NewFridgeManager(db *sql.DB) *FridgeManager {
	return &FridgeManager{db}
}

func (fm *FridgeManager) FindAll(ctx context.Context) ([]Fridge, error) {
	const op = apierror.Op("models.FridgeManager.FindAll")
	r := resolveRunner(ctx, fm.db)
	rows, err := r.QueryContext(ctx, `SELECT id, name, description, min_temp, max_temp, alerts_enabled FROM fridges`)
	if err != nil {
		return nil, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"failed to retrieve fridges",
			op,
		)
	}

	var fridges []Fridge
	for rows.Next() {
		var f Fridge
		err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.Description,
			&f.MinTemp,
			&f.MaxTemp,
			&f.AlertsEnabled,
		)
		if err != nil {
			return nil, apierror.Wrap(
				err,
				apierror.CodeDatabase,
				"failed to scan fridge row",
				op,
			)
		}
		fridges = append(fridges, f)
	}
	if err := rows.Err(); err != nil {
		return nil, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"error occurred while iterating over fridge rows",
			op,
		)
	}
	return fridges, nil
}

func (fm *FridgeManager) FindOneByID(ctx context.Context, id int64) (Fridge, error) {
	const op = apierror.Op("models.FridgeManager.FindOneByID")
	r := resolveRunner(ctx, fm.db)
	row := r.QueryRowContext(ctx, `SELECT id, name, description, min_temp, max_temp, alerts_enabled FROM fridges WHERE id = ?`, id)

	var f Fridge
	err := row.Scan(
		&f.ID,
		&f.Name,
		&f.Description,
		&f.MinTemp,
		&f.MaxTemp,
		&f.AlertsEnabled,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return f, apierror.New(
			apierror.CodeRecordNotFound,
			fmt.Sprintf("no fridge found with id %d", id),
			op,
		)
	} else if err != nil {
		return f, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"failed to retrieve fridge",
			op,
		)
	}
	return f, nil
}

func (fm *FridgeManager) InsertOne(ctx context.Context, fridge Fridge) (Fridge, error) {
	const op = apierror.Op("models.FridgeManager.InsertOne")
	var newFridge Fridge
	err := requireTxn(ctx).
		QueryRowContext(
			ctx,
			`INSERT INTO fridges(name, description, min_temp, max_temp, alerts_enabled) VALUES(?, ?, ?, ?, ?)
				RETURNING id, name, description, min_temp, max_temp, alerts_enabled`,
			fridge.Name,
			fridge.Description,
			fridge.MinTemp,
			fridge.MaxTemp,
			fridge.AlertsEnabled,
		).
		Scan(
			&newFridge.ID,
			&newFridge.Name,
			&newFridge.Description,
			&newFridge.MinTemp,
			&newFridge.MaxTemp,
			&newFridge.AlertsEnabled,
		)
	if err != nil {
		return newFridge, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"failed to insert fridge row",
			op,
		)
	}
	return newFridge, nil
}

type PartialFridge struct {
	Name          string
	Description   *string
	MinTemp       *float64
	MaxTemp       *float64
	AlertsEnabled *bool
}

func (fm *FridgeManager) UpdateOne(ctx context.Context, id int64, fridge PartialFridge) (Fridge, error) {
	const op = apierror.Op("models.FridgeManager.UpdateOne")
	var fields []string
	var args []any
	if fridge.Name != "" {
		fields = append(fields, "name")
		args = append(args, fridge.Name)
	}
	if fridge.Description != nil {
		fields = append(fields, "description")
		args = append(args, *fridge.Description)
	}
	if fridge.MinTemp != nil {
		fields = append(fields, "min_temp")
		args = append(args, *fridge.MinTemp)
	}
	if fridge.MaxTemp != nil {
		fields = append(fields, "max_temp")
		args = append(args, *fridge.MaxTemp)
	}
	if fridge.AlertsEnabled != nil {
		fields = append(fields, "alerts_enabled")
		args = append(args, *fridge.AlertsEnabled)
	}
	// If nothing to update just fetch and return the fridge
	if len(args) == 0 {
		return fm.FindOneByID(ctx, id)
	}

	var query strings.Builder
	query.WriteString("UPDATE fridges SET ")
	for i, field := range fields {
		if i > 0 {
			query.WriteString(", ")
		}
		query.WriteString(field)
		query.WriteString(" = ?")
	}
	query.WriteString(" WHERE id = ? RETURNING id, name, description, min_temp, max_temp, alerts_enabled")
	args = append(args, id)

	var newFridge Fridge
	err := requireTxn(ctx).
		QueryRowContext(ctx, query.String(), args...).
		Scan(
			&newFridge.ID,
			&newFridge.Name,
			&newFridge.Description,
			&newFridge.MinTemp,
			&newFridge.MaxTemp,
			&newFridge.AlertsEnabled,
		)
	if err != nil {
		return newFridge, apierror.Wrap(
			err,
			apierror.CodeDatabase,
			"failed to update fridge row",
			op,
		)
	}
	return newFridge, nil
}
