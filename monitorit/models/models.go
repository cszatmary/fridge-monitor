package models

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

type txnKey struct{}

func ContextWithTxn(ctx context.Context, txn *sql.Tx) context.Context {
	return context.WithValue(ctx, txnKey{}, txn)
}

func txnFromContext(ctx context.Context) *sql.Tx {
	v := ctx.Value(txnKey{})
	if v == nil {
		return nil
	}
	txn, ok := v.(*sql.Tx)
	if !ok {
		panic("impossible: value is not of type *sql.Tx")
	}
	return txn
}

// runner represents functionality for working with a DB.
// It is a way to generalize sql.DB and sql.Tx.
type runner interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func resolveRunner(ctx context.Context, db *sql.DB) runner {
	txn := txnFromContext(ctx)
	if txn == nil {
		return db
	}
	return txn
}

func requireTxn(ctx context.Context) *sql.Tx {
	txn := txnFromContext(ctx)
	if txn == nil {
		panic("transaction is required but is not present")
	}
	return txn
}

const sqliteTimeFormat = "2006-01-02 15:04:05"

// Time is a time.Time that implements the sql.Scanner interface and
// can be used as a scan destination for datetimes stored in sqlite text columns.
type Time struct{ time.Time }

func (t *Time) Scan(value any) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unsupported source type for Time: %T", value)
	}
	parsed, err := time.Parse(sqliteTimeFormat, s)
	if err != nil {
		return fmt.Errorf("failed to parse scanned value as a Time: %s: %w", s, err)
	}
	t.Time = parsed
	return nil
}

func (t Time) Value() (driver.Value, error) {
	return t.Format(sqliteTimeFormat), nil
}
