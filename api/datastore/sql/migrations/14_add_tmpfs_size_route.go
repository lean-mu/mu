package migrations

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lean-mu/mu/api/datastore/sql/migratex"
)

func up14(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE routes ADD tmpfs_size int;")

	return err
}

func down14(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE routes DROP COLUMN tmpfs_size;")
	return err
}

func init() {
	Migrations = append(Migrations, &migratex.MigFields{
		VersionFunc: vfunc(14),
		UpFunc:      up14,
		DownFunc:    down14,
	})
}
