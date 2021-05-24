package migrations

import (
	"github.com/lean-mu/mu/api/datastore/sql/migratex"
)

// Migrations is the list of fn specific sql migrations to run
var Migrations []migratex.Migration

func vfunc(v int64) func() int64 { return func() int64 { return v } }
