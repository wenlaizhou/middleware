package middleware

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"time"
)

type Database struct {
	conn           *sql.DB
	dbName         string
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
		dbName:         dbName,
		timeoutSeconds: time.Duration(timeoutSecond) * time.Second,
	}, nil
}

// ? 代表参数
func (d Database) Query(sql string, params ...interface{}) ([]map[string]string, error) {
	result := []map[string]string{}
	timeoutContext, _ := context.WithTimeout(context.Background(), d.timeoutSeconds)
	rows, err := d.conn.QueryContext(timeoutContext, sql, params...)
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
func (d Database) Exec(sql string, params ...interface{}) (int64, int64, error) {
	timeoutContext, _ := context.WithTimeout(context.Background(), d.timeoutSeconds)
	rows, err := d.conn.ExecContext(timeoutContext, sql, params...)
	if err != nil {
		return -1, -1, err
	}
	rowsAffected, _ := rows.RowsAffected()
	lastInsertedId, _ := rows.LastInsertId()
	return rowsAffected, lastInsertedId, nil
}

type DatabaseSchema struct {
	Name   string                 `json:"name"`
	Tables map[string]TableSchema `json:"tables"`
}

type TableSchema struct {
	Name    string                  `json:"name"`
	Columns map[string]ColumnSchema `json:"columns"`
}

type ColumnSchema struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Comment  string `json:"comment"`
}

func (d Database) Schema() (DatabaseSchema, error) {
	res, err := d.Query("select table_name, column_name, data_type, column_comment from information_schema.columns where table_schema = ? order by table_name", d.dbName)
	if err != nil {
		return DatabaseSchema{}, err
	}
	result := DatabaseSchema{
		Name:   d.dbName,
		Tables: map[string]TableSchema{},
	}
	if len(res) <= 0 {
		return DatabaseSchema{}, errors.New("no privileges")
	}
	for _, row := range res {
		tableName, has := row["table_name"]
		if !has {
			continue
		}
		t, has := result.Tables[tableName]
		if has {
			t.Columns[row["column_name"]] = ColumnSchema{
				Name:     row["column_name"],
				DataType: row["data_type"],
				Comment:  row["column_comment"],
			}
		} else {
			result.Tables[tableName] = TableSchema{
				Name: tableName,
				Columns: map[string]ColumnSchema{
					row["column_name"]: ColumnSchema{
						Name:     row["column_name"],
						DataType: row["data_type"],
						Comment:  row["column_comment"],
					}},
			}
		}
	}
	return result, nil
}

func (d Database) Close() {
	d.conn.Close()
}
