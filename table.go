package main

import (
	"fmt"
	"sort"
	"strings"
)

func getTables(e *DbExplorer) []string {
	tables := make([]string, 0, len(e.tables))
	for table := range e.tables {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	return tables
}

func getTableRecords(e *DbExplorer, table string) ([]map[string]any, error) {
	safeTableName, err := getSafeTableName(e, table)
	if err != nil {
		return nil, err
	}

	rows, err := e.db.Query("SELECT * FROM " + safeTableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	records := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		dest := make([]any, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		record := make(map[string]any, len(columns))
		for i, col := range columns {
			value := values[i]
			switch typedValue := value.(type) {
			case []byte:
				record[col] = string(typedValue)
			default:
				record[col] = typedValue
			}
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func getSafeTableName(e *DbExplorer, table string) (string, error) {
	if _, ok := e.tables[table]; !ok {
		return "", fmt.Errorf("unknown table")
	}

	return "`" + strings.ReplaceAll(table, "`", "``") + "`", nil
}
