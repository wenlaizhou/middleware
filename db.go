package middleware

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

// dsn格式: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]

func PooledConnection(maxConnections int) *sql.DB {
	// Opening a driver typically will not attempt to connect to the database.
	db, err := sql.Open("mysql", "[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]")
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(maxConnections)
	db.SetMaxOpenConns(maxConnections)
	return db
}
