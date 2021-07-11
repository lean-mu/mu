package migrations

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lean-mu/mu/api/datastore/sql/migratex"
)

func up6(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE apps ADD updated_at VARCHAR(256);")
	return err
}

func down6(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE apps DROP COLUMN updated_at;")
	return err
}

func init() {
	Migrations = append(Migrations, &migratex.MigFields{
		VersionFunc: vfunc(6),
		UpFunc:      up6,
		DownFunc:    down6,
	})
}
