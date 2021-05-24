package migrations

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lean-mu/mu/api/datastore/sql/migratex"
)

func up9(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE routes ADD annotations TEXT;")

	return err
}

func down9(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE routes DROP COLUMN annotations;")
	return err
}

func init() {
	Migrations = append(Migrations, &migratex.MigFields{
		VersionFunc: vfunc(9),
		UpFunc:      up9,
		DownFunc:    down9,
	})
}
