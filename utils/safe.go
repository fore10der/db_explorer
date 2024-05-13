package utils

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type TableFieldInfo struct {
	Name      string
	TypeName  string
	Required  bool
	IsPrimary bool
}

func GetSafeTableName(tables map[string]struct{}, table string) (string, error) {
	if _, ok := tables[table]; !ok {
		return "", fmt.Errorf("unknown table")
	}

	return "`" + strings.ReplaceAll(table, "`", "``") + "`", nil
}

func GetInsertFieldsAndValues(data map[string]any, fields []TableFieldInfo, checkRequired bool) ([]string, []string, error) {
	columns := make([]string, 0, len(fields))
	values := make([]string, 0, len(fields))

	for _, field := range fields {
		value, exists := data[field.Name]
		if !exists {
			if checkRequired && field.Required {
				return nil, nil, fmt.Errorf("field %s is required", field.Name)
			}
			continue
		}

		if field.IsPrimary {
			continue
		}

		// неизвестные поля автоматически игнорируются, потому что
		// идём только по схеме таблицы
		columns = append(columns, field.Name)
		values = append(values, formatSQLValue(value))
	}

	return columns, values, nil
}

func formatSQLValue(value any) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		return "'" + escapeSQLString(v) + "'"
	case []byte:
		return "'" + escapeSQLString(string(v)) + "'"
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return "'" + escapeSQLString(fmt.Sprintf("%v", v)) + "'"
	}
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
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
