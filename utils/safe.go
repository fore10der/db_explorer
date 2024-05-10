package utils

import (
	"database/sql"
	"fmt"
	"strings"
)

func GetSafeTableName(tables map[string]struct{}, table string) (string, error) {
	if _, ok := tables[table]; !ok {
		return "", fmt.Errorf("unknown table")
	}

	return "`" + strings.ReplaceAll(table, "`", "``") + "`", nil
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
