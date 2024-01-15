package database

import (
	"context"
	"database/sql"
)

type ctxKey struct{}

type DB interface {
	QueryRow(string, ...any) *sql.Row
	Exec(string, ...any) (sql.Result, error)
}

func AddToContext(ctx context.Context, db DB) context.Context {
	return context.WithValue(ctx, ctxKey{}, db)
}

func FromContext(ctx context.Context) (DB, bool) {
	val, ok := ctx.Value(ctxKey{}).(DB)
	return val, ok
}
