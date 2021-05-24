package migrations

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lean-mu/mu/api/datastore/sql/migratex"
)

func up2(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE calls ADD stats text;")
	return err
}

func down2(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "ALTER TABLE calls DROP COLUMN stats;")
	return err
}

func init() {
	Migrations = append(Migrations, &migratex.MigFields{
		VersionFunc: vfunc(2),
		UpFunc:      up2,
		DownFunc:    down2,
	})
}
