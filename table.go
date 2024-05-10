package main

import (
	"db_explorer/utils"
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

	rows, err := e.db.Query("SELECT * FROM "+safeTableName+" WHERE id = ? LIMIT 1", id)
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

	if !rows.Next() {
		return nil, fmt.Errorf("record not found")
	}

	return utils.ScanRowToRecord(rows, columns, columnTypes)
}

func deleteTableRecord(e *DbExplorer, table string, id int64) (int64, error) {
	safeTableName, err := utils.GetSafeTableName(e.tables, table)
	if err != nil {
		return 0, err
	}

	primaryKey, err := utils.GetPrimaryKeyColumn(e.db, safeTableName)
	if err != nil {
		return 0, err
	}

	safePrimaryKey := "`" + strings.ReplaceAll(primaryKey, "`", "``") + "`"
	result, err := e.db.Exec("DELETE FROM "+safeTableName+" WHERE "+safePrimaryKey+" = ?", id)
	if err != nil {
		return 0, err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return deleted, nil
}
