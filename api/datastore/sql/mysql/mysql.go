package mysql

import (
	"net/url"
	"strings"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/lean-mu/mu/api/datastore/sql/dbhelper"
)

type mysqlHelper int

func (mysqlHelper) Supports(scheme string) bool {
	return scheme == "mysql"
}

func (mysqlHelper) PreConnect(url *url.URL) (string, error) {
	return strings.TrimPrefix(url.String(), url.Scheme+"://"), nil
}

func (mysqlHelper) PostCreate(db *sqlx.DB) (*sqlx.DB, error) {
	return db, nil

}
func (mysqlHelper) CheckTableExists(tx *sqlx.Tx, table string) (bool, error) {
	query := tx.Rebind(`SELECT count(*)
	FROM information_schema.TABLES
	WHERE TABLE_NAME = ?`)

	row := tx.QueryRow(query, table)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	exists := count > 0
	return exists, nil
}

func (mysqlHelper) String() string {
	return "mysql"
}

func (mysqlHelper) IsDuplicateKeyError(err error) bool {
	switch mErr := err.(type) {
	case *mysql.MySQLError:
		if mErr.Number == 1062 {
			return true
		}
	}
	return false
}

func init() {
	dbhelper.Register(mysqlHelper(0))
}
