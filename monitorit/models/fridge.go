package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cszatmary/fridge-monitor/monitorit/lib/apierror"
)

type Fridge struct {
	ID          int64
	Name        string
	Description string
	MinTemp     float64
	MaxTemp     float64
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
	rows, err := r.QueryContext(ctx, `SELECT id, name, description, min_temp, max_temp FROM fridges`)
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
	row := r.QueryRowContext(ctx, `SELECT id, name, description, min_temp, max_temp FROM fridges WHERE id = ?`, id)

	var f Fridge
	err := row.Scan(
		&f.ID,
		&f.Name,
		&f.Description,
		&f.MinTemp,
		&f.MaxTemp,
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
