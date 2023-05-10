package db

import (
	"database/sql"
	"encoding/json"
)

func ConvertSQLRowsToJSON(r *sql.Rows) ([]byte, error) {
	var finalRows []interface{}
	var res []byte

	columnTypes, err := r.ColumnTypes()

	if err != nil {
		return nil, err
	}

	count := len(columnTypes)

	for r.Next() {
		scanArgs := make([]interface{}, count)
		for i, v := range columnTypes {
			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
				scanArgs[i] = new(sql.NullString)
				break
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
				break
			case "INT":
				scanArgs[i] = new(sql.NullInt64)
				break
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err = r.Scan(scanArgs...)

		if err != nil {
			return nil, err
		}

		masterData := map[string]interface{}{}

		for i, v := range columnTypes {

			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				masterData[v.Name()] = z.Bool
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				masterData[v.Name()] = z.String
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				masterData[v.Name()] = z.Int64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				masterData[v.Name()] = z.Int32
				continue
			}

			masterData[v.Name()] = scanArgs[i]
		}
		finalRows = append(finalRows, masterData)
	}
	if len(finalRows) == 1 {
		res, err = json.Marshal(finalRows[0])
		if err != nil {
			return nil, err
		}
	} else {
		res, err = json.Marshal(finalRows)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
