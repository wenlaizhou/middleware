package middleware

import (
	"database/sql"
)

// dsn格式: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// 使用时需添加驱动: _ "github.com/go-sql-driver/mysql"
func PooledConnection(driver string, dsn string, maxConnections int) (*sql.DB, error) {
	// Opening a driver typically will not attempt to connect to the database.
	db, err := sql.Open(driver, dsn)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		return nil, err
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(maxConnections)
	db.SetMaxOpenConns(maxConnections)
	return db, nil
}
