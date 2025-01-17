package migrations

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lean-mu/mu/api/datastore/sql/migratex"
)

func up16(ctx context.Context, tx *sqlx.Tx) error {
	createQuery := `CREATE TABLE IF NOT EXISTS fns (
	id varchar(256) NOT NULL PRIMARY KEY,
	name varchar(256) NOT NULL,
	app_id varchar(256) NOT NULL,
	image varchar(256) NOT NULL,
	format varchar(16) NOT NULL,
	memory int NOT NULL,
	timeout int NOT NULL,
	idle_timeout int NOT NULL,
	config text NOT NULL,
	annotations text NOT NULL,
	created_at varchar(256) NOT NULL,
	updated_at varchar(256) NOT NULL,
    CONSTRAINT name_app_id_unique UNIQUE (app_id, name)
);`
	_, err := tx.ExecContext(ctx, createQuery)
	return err
}

func down16(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP TABLE fns;")
	return err
}

func init() {
	Migrations = append(Migrations, &migratex.MigFields{
		VersionFunc: vfunc(16),
		UpFunc:      up16,
		DownFunc:    down16,
	})
}
