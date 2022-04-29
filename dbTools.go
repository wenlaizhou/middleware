package middleware

import "database/sql"

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
			} else {
				row[columns[i]] = string(data[i].([]byte))
			}
		}
		result = append(result, row)
	}
	rows.Close()
	return result, nil
}
