package middleware

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"time"
)

type Database struct {
	conn *sql.DB
	cfg  mysql.Config
}

func DbPool(addr string, user string, password string, dbName string, maxConnections int, timeoutSecond int) (Database, error) {
	cfg := mysql.Config{
		User:                 user,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 addr,
		DBName:               dbName,
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return Database{}, err
	}
	db.SetConnMaxLifetime(time.Duration(timeoutSecond) * time.Second)
	db.SetMaxIdleConns(maxConnections)
	db.SetMaxIdleConns(maxConnections)
	// Confirm a successful connection.
	if err := db.Ping(); err != nil {
		return Database{}, err
	}
	return Database{
		conn: db,
		cfg:  cfg,
	}, nil
}

// ? 代表参数
func (thisSelf Database) Query(sql string, params ...interface{}) ([]map[string]string, error) {
	result := []map[string]string{}
	rows, err := thisSelf.conn.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	columns, _ := rows.Columns()
	for rows.Next() {
		data := make([]string, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i, _ := range data {
			columnPointers[i] = &data[i]
		}
		err = rows.Scan(columnPointers...)
		if err != nil {
			return nil, err
		}
		row := map[string]string{}
		for i, _ := range columns {
			row[columns[i]] = data[i]
		}
		result = append(result, row)
	}
	rows.Close()
	return result, nil
}

// ? 代表参数
func (thisSelf Database) Exec(sql string, params ...interface{}) (int64, int64, error) {
	rows, err := thisSelf.conn.Exec(sql, params...)
	if err != nil {
		return -1, -1, err
	}
	rowsAffected, _ := rows.RowsAffected()
	lastInsertedId, _ := rows.LastInsertId()
	return rowsAffected, lastInsertedId, nil
}

func (thisSelf Database) Close() {
	thisSelf.conn.Close()
}
