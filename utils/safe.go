package utils

import (
	"database/sql"
	"encoding/json"
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

func GetDatabaseFieldsAndValues(data map[string]any, fields []TableFieldInfo, checkRequired bool) ([]string, []string, error) {
	columns := make([]string, 0, len(fields))
	values := make([]string, 0, len(fields))
	fmt.Println(data)

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

		if err := validateFieldValueByDBType(value, field); err != nil {
			return nil, nil, err
		}

		// неизвестные поля автоматически игнорируются, потому что
		// идём только по схеме таблицы
		columns = append(columns, field.Name)
		values = append(values, formatSQLValue(value))
	}

	return columns, values, nil
}

func validateFieldValueByDBType(value any, field TableFieldInfo) error {
	if value == nil {
		if field.Required {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}
		return nil
	}

	expectedKind := getExpectedValueKind(field.TypeName)
	switch expectedKind {
	case "number":
		if !isNumberValue(value) {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}
	case "string":
		if !isStringValue(value) {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}
	case "boolean":
		if !isBooleanValue(value) {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}
	}

	return nil
}

func getExpectedValueKind(typeName string) string {
	normalized := strings.ToLower(strings.TrimSpace(typeName))
	compact := strings.ReplaceAll(normalized, " ", "")

	if strings.HasPrefix(compact, "bool") || strings.HasPrefix(compact, "boolean") || strings.HasPrefix(compact, "tinyint(1)") || strings.HasPrefix(compact, "bit(1)") {
		return "boolean"
	}

	base := normalized
	if idx := strings.Index(base, "("); idx != -1 {
		base = base[:idx]
	}
	if idx := strings.Index(base, " "); idx != -1 {
		base = base[:idx]
	}

	switch base {
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "decimal", "dec", "numeric", "fixed", "float", "double", "real", "bit":
		return "number"
	case "char", "varchar", "text", "tinytext", "mediumtext", "longtext", "enum", "set", "json", "date", "datetime", "timestamp", "time", "year", "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		return "string"
	default:
		return "string"
	}
}

func isNumberValue(value any) bool {
	switch v := value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	case json.Number:
		_, err := v.Float64()
		return err == nil
	default:
		return false
	}
}

func isStringValue(value any) bool {
	switch value.(type) {
	case string, []byte:
		return true
	default:
		return false
	}
}

func isBooleanValue(value any) bool {
	_, ok := value.(bool)
	return ok
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
