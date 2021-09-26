package middleware

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func PooledConnection() {
	// Opening a driver typically will not attempt to connect to the database.
	db, err := sql.Open("driver-name", "database=test1")
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
}
