package driver

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLClient struct {
	conn *sql.DB
}

func NewMySQLClient(user string, password string, host string, port string, dbName string) MySQLClient {
	db := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true",
		user, password, host, port, dbName,
	)

	conn, err := sql.Open("mysql", db)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	return MySQLClient{
		conn: conn,
	}
}

func (c MySQLClient) FindLatestMigrationVersion() (int, error) {
	stmtMigrationVersion, err := c.conn.Prepare("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1")
	if err != nil {
		return 0, err
	}
	defer stmtMigrationVersion.Close()

	var migrationVersion int
	if err = stmtMigrationVersion.QueryRow().Scan(&migrationVersion); err != nil {
		return 0, err
	}

	return migrationVersion, nil
}
