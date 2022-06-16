package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
)

var jsonType = reflect.TypeOf(json.RawMessage{})

// ScanRows rows扫描
func ScanRows(rows *sql.Rows) ([]map[string]string, error) {
	result := []map[string]string{}
	columns, _ := rows.Columns()
	for rows.Next() {
		data := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i, _ := range data {
			columnPointers[i] = &data[i]
		}
		err := rows.Scan(columnPointers...)
		if err != nil {
			return nil, err
		}
		row := map[string]string{}
		for i, _ := range columns {
			if data[i] == nil {
				row[columns[i]] = ""
				continue
			}
			row[columns[i]] = convertVal(data[i])
		}
		result = append(result, row)
	}
	rows.Close()
	return result, nil
}

func convertVal(param interface{}) string {
	dataVal := reflect.ValueOf(param)
	switch dataVal.Kind() {
	case reflect.Ptr:
		// indirect pointers
		if dataVal.IsNil() {
			return ""
		} else {
			return convertVal(dataVal.Elem().Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", dataVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%v", dataVal.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.4f", dataVal.Float())
	case reflect.Bool:
		return fmt.Sprintf("%v", dataVal.Bool())
	case reflect.Slice:
		switch t := dataVal.Type(); {
		case t == jsonType:
			return fmt.Sprintf("%v", t)
		case t.Elem().Kind() == reflect.Uint8:
			return fmt.Sprintf("%v", string(dataVal.Bytes()))
		default:
			return fmt.Sprintf("unsupported type %T, a slice of %s", param, t.Elem().Kind())
		}
	case reflect.String:
		return dataVal.String()
	}
	return fmt.Sprintf("unsupported type %T, a %s", param, dataVal.Kind())
}
