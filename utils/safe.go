package utils

import (
	"database/sql"
	"fmt"
	"strings"
)

type TableFieldInfo struct {
	Name     string
	TypeName string
	Required bool
}

func GetSafeTableName(tables map[string]struct{}, table string) (string, error) {
	if _, ok := tables[table]; !ok {
		return "", fmt.Errorf("unknown table")
	}

	return "`" + strings.ReplaceAll(table, "`", "``") + "`", nil
}

func GetInsertFieldsAndValues(data map[string]any, fields []TableFieldInfo, checkRequired bool) ([]string, []any, error) {
	columns := make([]string, 0, len(fields))
	values := make([]any, 0, len(fields))

	for _, field := range fields {
		value, exists := data[field.Name]
		if !exists {
			if checkRequired && field.Required {
				return nil, nil, fmt.Errorf("field %s is required", field.Name)
			}
			continue
		}

		// неизвестные поля автоматически игнорируются, потому что
		// идём только по схеме таблицы
		columns = append(columns, field.Name)
		values = append(values, value)
	}

	return columns, values, nil
}

func GetPrimaryKeyColumn(db *sql.DB, safeTableName string) (string, error) {
	rows, err := db.Query("SHOW FULL COLUMNS FROM " + safeTableName)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			field      string
			typeName   string
			collation  sql.NullString
			null       string
			key        string
			def        sql.NullString
			extra      string
			privileges string
			comment    string
		)

		if err := rows.Scan(&field, &typeName, &collation, &null, &key, &def, &extra, &privileges, &comment); err != nil {
			return "", err
		}
		if key == "PRI" {
			return field, nil
		}
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return "id", nil
}
