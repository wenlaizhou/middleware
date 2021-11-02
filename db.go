package middleware

import (
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"time"
)

type Database struct {
	conn           *sql.DB
	cfg            mysql.Config
	timeoutSeconds time.Duration
}

// 数据库连接池 最长默认存活时间
// 可进行设置
var DefaultConnectionLifeSeconds = time.Duration(60*10) * time.Second

// 创建数据库连接池
// addr: 数据库链接地址, ip:port
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
	db.SetConnMaxLifetime(DefaultConnectionLifeSeconds)
	db.SetConnMaxIdleTime(DefaultConnectionLifeSeconds)
	db.SetMaxIdleConns(maxConnections)
	db.SetMaxOpenConns(maxConnections)

	// Confirm a successful connection.
	timeoutContext, _ := context.WithTimeout(context.Background(), time.Duration(timeoutSecond)*time.Second)
	if err := db.PingContext(timeoutContext); err != nil {
		return Database{}, err
	}

	return Database{
		conn:           db,
		cfg:            cfg,
		timeoutSeconds: time.Duration(timeoutSecond) * time.Second,
	}, nil
}

// ? 代表参数
func (thisSelf Database) Query(sql string, params ...interface{}) ([]map[string]string, error) {
	result := []map[string]string{}
	timeoutContext, _ := context.WithTimeout(context.Background(), thisSelf.timeoutSeconds)
	rows, err := thisSelf.conn.QueryContext(timeoutContext, sql, params...)
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
	timeoutContext, _ := context.WithTimeout(context.Background(), thisSelf.timeoutSeconds)
	rows, err := thisSelf.conn.ExecContext(timeoutContext, sql, params...)
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
