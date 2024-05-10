package utils

import (
	"database/sql"
	"math"
	"strconv"
	"strings"
)

type ScannableRow interface {
	Scan(dest ...any) error
}

func ScanRowToRecord(row ScannableRow, columns []string, columnTypes []*sql.ColumnType) (map[string]any, error) {
	values := make([]any, len(columns))
	dest := make([]any, len(columns))
	for i := range values {
		dest[i] = &values[i]
	}

	if err := row.Scan(dest...); err != nil {
		return nil, err
	}

	mappedRecord := make(map[string]any, len(columns))
	for i, col := range columns {
		value := values[i]
		dbTypeName := ""
		if i < len(columnTypes) {
			dbTypeName = columnTypes[i].DatabaseTypeName()
		}
		mappedRecord[col] = NormalizeRecordField(value, dbTypeName)
	}

	return mappedRecord, nil
}

func NormalizeRecordField(value any, dbTypeName string) any {
	if value == nil {
		return nil
	}

	switch typedValue := value.(type) {
	case []byte:
		return normalizeStringByDBType(string(typedValue), dbTypeName)
	default:
		return typedValue
	}
}

func normalizeStringByDBType(value string, dbTypeName string) any {
	typeName := strings.ToUpper(strings.TrimSpace(dbTypeName))

	if isBoolDBType(typeName) {
		if value == "1" {
			return true
		}
		if value == "0" {
			return false
		}
		if parsed, err := strconv.ParseBool(strings.ToLower(value)); err == nil {
			return parsed
		}
		return value
	}

	if isIntDBType(typeName) {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
		if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
			if parsed <= math.MaxInt64 {
				return int64(parsed)
			}
			return parsed
		}
		return value
	}

	if isFloatDBType(typeName) {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
		return value
	}

	return value
}

func isBoolDBType(typeName string) bool {
	switch typeName {
	case "BOOL", "BOOLEAN", "BIT":
		return true
	default:
		return false
	}
}

func isIntDBType(typeName string) bool {
	switch typeName {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT":
		return true
	default:
		return false
	}
}

func isFloatDBType(typeName string) bool {
	switch typeName {
	case "FLOAT", "DOUBLE", "REAL", "DECIMAL", "NUMERIC":
		return true
	default:
		return false
	}
}
