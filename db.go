package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"html"
	"regexp"
	"strings"
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
		data := make([]interface{}, len(columns))
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
			if data[i] == nil {
				row[columns[i]] = ""
			} else {
				row[columns[i]] = string(data[i].([]byte))
			}
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

var sqlCheckReg = regexp.MustCompile("where")

func SqlParamCheck(p string) bool {
	if html.EscapeString(p) != p {
		return false
	}
	if sqlCheckReg.MatchString(p) {
		return false
	}
	return true
}

func SqlParamIsTp(p string) bool {
	checker := strings.ToLower(p)
	reg := regexp.MustCompile("(insert)|(update)|(delete)|(alter)|(modify)")
	return reg.MatchString(checker)
}

func RegisterDbHandler(d Database, prefix string) []SwaggerPath {

	dbHandlerLogger := GetLogger(fmt.Sprintf("db-%s", d.dbName))

	if !strings.HasPrefix(prefix, "/") {
		prefix = fmt.Sprintf("/%s", prefix)
	}

	schemaSwagger := SwaggerBuildPath(fmt.Sprintf("%s/schema", prefix), d.dbName, "get", "db schema")
	RegisterHandler(fmt.Sprintf("%s/schema", prefix), func(c Context) {
		res, err := d.Schema()
		if err != nil {
			c.ApiResponse(-1, err.Error(), nil)
			return
		}
		c.ApiResponse(0, "", res)
		return
	})

	selectSwagger := SwaggerBuildPath(fmt.Sprintf("%s/select/{table}", prefix), d.dbName, "get", "select from table")
	selectSwagger.AddParameter(SwaggerParameter{
		Name:        "table",
		Description: "table name",
		In:          "path",
		Required:    true,
	})
	selectSwagger.AddParameter(SwaggerParameter{
		Name:        "limit",
		Description: "limit",
		In:          "query",
		Required:    false,
	})
	selectSwagger.AddParameter(SwaggerParameter{
		Name:        "orderBy",
		Description: "order by ... desc",
		In:          "query",
		Required:    false,
	})
	RegisterHandler(fmt.Sprintf("%s/select/{table}", prefix), func(c Context) {
		table := c.GetPathParam("table")
		if !SqlParamCheck(table) {
			c.ApiResponse(-1, "", nil)
			return
		}

		limitSql := ""
		if limit := c.GetQueryParam("limit"); len(limit) > 0 && SqlParamCheck(limit) {
			limit = strings.TrimSpace(limit)
			if SqlParamIsTp(limit) {
				c.ApiResponse(-1, "limit has tp sql", nil)
				return
			}
			limitSql = fmt.Sprintf("limit %v", limit)
		}
		orderBySql := ""
		if orderBy := c.GetQueryParam("orderBy"); len(orderBy) > 0 && SqlParamCheck(orderBy) {
			orderBy = strings.TrimSpace(orderBy)
			if SqlParamIsTp(orderBy) {
				c.ApiResponse(-1, "orderBy has tp sql", nil)
				return
			}
			orderBySql = fmt.Sprintf("order by %v", orderBy)
		}
		selectSql := strings.TrimSpace(fmt.Sprintf("select * from %s %s %s", table, orderBySql, limitSql))
		dbHandlerLogger.InfoF("sql: %s", selectSql)
		res, err := d.Query(selectSql)
		if err != nil {
			c.ApiResponse(-1, err.Error(), nil)
			return
		}
		c.ApiResponse(0, "", res)
		return
	})

	insertSwagger := SwaggerBuildPath(fmt.Sprintf("%s/insert/{table}", prefix), d.dbName, "post", "insert into table")
	insertSwagger.AddParameter(SwaggerParameter{
		Name:        "table",
		Description: "table name",
		In:          "path",
		Required:    true,
	})
	insertSwagger.AddParameter(SwaggerParameter{
		Name:        "json",
		Default:     "{}",
		Description: "对象类型数据",
		In:          "body",
		Required:    true,
	})
	RegisterHandler(fmt.Sprintf("%s/insert/{table}", prefix), func(c Context) {
		table := c.GetPathParam("table")
		if !SqlParamCheck(table) {
			c.ApiResponse(-1, "", nil)
			return
		}
		params, err := c.GetJSON()
		if err != nil {
			c.ApiResponse(-1, err.Error(), nil)
			return
		}
		if len(params) <= 0 {
			c.ApiResponse(-1, "invalid params", nil)
			return
		}
		insertSql := fmt.Sprintf("insert into %v (", table)
		values := "values ("
		sqlParams := []interface{}{}
		first := true
		for k, v := range params {
			if !SqlParamCheck(k) {
				c.ApiResponse(-1, fmt.Sprintf("invalid Key : %s", k), nil)
				return
			}
			if first {
				insertSql = fmt.Sprintf("%s %s", insertSql, k)
				values = fmt.Sprintf("%s ?", values)
			} else {
				insertSql = fmt.Sprintf("%s, %s", insertSql, k)
				values = fmt.Sprintf("%s, ?", values)
			}
			sqlParams = append(sqlParams, v)
			first = false
		}
		//insertSql = insertSql[:len(insertSql)-2]
		insertSql = fmt.Sprintf("%s )", insertSql)

		//values = values[:len(values)-2]
		values = fmt.Sprintf("%s )", values)

		insertSql = fmt.Sprintf("%s %s", insertSql, values)

		dbHandlerLogger.InfoF("sql: %s, params: %v", insertSql, sqlParams)
		c.ApiResponse(0, insertSql, sqlParams)
		return
		//d.Exec(insertSql, sqlParams...)

	})

	updateSwagger := SwaggerBuildPath(fmt.Sprintf("%s/update/{table}", prefix), d.dbName, "post", "update table")
	updateSwagger.AddParameter(SwaggerParameter{
		Name:        "table",
		Description: "table name",
		In:          "path",
		Required:    true,
	})
	updateSwagger.AddParameter(SwaggerParameter{
		Name: "json",
		Default: `{
  "id" : 1
}`,
		Description: "必须具有id字段进行数据定位",
		In:          "body",
		Required:    true,
	})
	RegisterHandler(fmt.Sprintf("%s/update/{table}", prefix), func(c Context) {
		table := c.GetPathParam("table")
		if !SqlParamCheck(table) {
			c.ApiResponse(-1, "", nil)
			return
		}
		params, err := c.GetJSON()
		if err != nil {
			c.ApiResponse(-1, err.Error(), nil)
			return
		}
		if len(params) <= 0 {
			c.ApiResponse(-1, "invalid params", nil)
			return
		}
		updateSql := fmt.Sprintf("update %v set (", table)
		sqlParams := []interface{}{}
		id, hasId := params["id"]
		if !hasId {
			c.ApiResponse(-1, "no id", nil)
			return
		}
		first := true
		for k, v := range params {
			if !SqlParamCheck(k) {
				c.ApiResponse(-1, fmt.Sprintf("invalid Param : %s", k), nil)
				return
			}
			if k == "id" {
				continue
			}
			if first {
				updateSql = fmt.Sprintf("%s %s = ?", updateSql, k)
			} else {
				updateSql = fmt.Sprintf("%s, %s = ?", updateSql, k)
			}
			sqlParams = append(sqlParams, v)
			first = false
		}

		updateSql = fmt.Sprintf("%s ) where id = ?", updateSql)
		sqlParams = append(sqlParams, id)
		dbHandlerLogger.InfoF("sql: %s, params: %v", updateSql, sqlParams)
		c.ApiResponse(0, updateSql, sqlParams)
		return
	})

	deleteSwagger := SwaggerBuildPath(fmt.Sprintf("%s/delete/{table}", prefix), d.dbName, "post", "delete from table")
	deleteSwagger.AddParameter(SwaggerParameter{
		Name:        "table",
		Description: "table name",
		In:          "path",
		Required:    true,
	})
	deleteSwagger.AddParameter(SwaggerParameter{
		Name: "json",
		Default: `{
  "id" : 1
}`,
		Description: "必须具有id字段进行数据定位",
		In:          "body",
		Required:    true,
	})
	RegisterHandler(fmt.Sprintf("%s/delete/{table}", prefix), func(c Context) {
		table := c.GetPathParam("table")
		if !SqlParamCheck(table) {
			c.ApiResponse(-1, "", nil)
			return
		}
		params, err := c.GetJSON()
		if err != nil {
			c.ApiResponse(-1, err.Error(), nil)
			return
		}
		if len(params) <= 0 {
			c.ApiResponse(-1, "invalid params", nil)
			return
		}
		id, hasId := params["id"]
		if !hasId {
			c.ApiResponse(-1, "no id", nil)
			return
		}
		deleteSql := fmt.Sprintf("delete from %s where id = ?", table)
		dbHandlerLogger.InfoF("sql: %s, params: %v", deleteSql, id)
		c.ApiResponse(0, deleteSql, id)
		return
	})

	return []SwaggerPath{schemaSwagger, selectSwagger, insertSwagger, updateSwagger, deleteSwagger}
}
