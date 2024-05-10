package main

import (
	"db_explorer/utils"
	"sort"
)

func getTables(e *DbExplorer) []string {
	tables := make([]string, 0, len(e.tables))
	for table := range e.tables {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	return tables
}

func getTableRecords(e *DbExplorer, table string, limit, offset int) ([]map[string]any, error) {
	safeTableName, err := utils.GetSafeTableName(e.tables, table)
	if err != nil {
		return nil, err
	}

	rows, err := e.db.Query("SELECT * FROM "+safeTableName+" LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	records := make([]map[string]any, 0)
	for rows.Next() {
		record, err := utils.ScanRowToRecord(rows, columns, columnTypes)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func getTableRecord(e *DbExplorer, table string, id int64) (map[string]any, error) {
	safeTableName, err := utils.GetSafeTableName(e.tables, table)
	if err != nil {
		return nil, err
	}

	metaRows, err := e.db.Query("SELECT * FROM " + safeTableName + " LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer metaRows.Close()

	columns, err := metaRows.Columns()
	if err != nil {
		return nil, err
	}

	columnTypes, err := metaRows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	row := e.db.QueryRow("SELECT * FROM "+safeTableName+" WHERE id = ?", id)
	return utils.ScanRowToRecord(row, columns, columnTypes)
}
