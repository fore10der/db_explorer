package main

import (
	"database/sql"
	"db_explorer/utils"
	"fmt"
	"sort"
	"strings"
)

func findTablePrimaryKeyField(fields []utils.TableFieldInfo) (utils.TableFieldInfo, error) {
	for _, field := range fields {
		if field.IsPrimary {
			return field, nil
		}
	}
	return utils.TableFieldInfo{}, fmt.Errorf("no primary key field found")
}

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

	primaryKey, err := utils.GetPrimaryKeyColumn(e.db, safeTableName)
	if err != nil {
		return nil, err
	}

	safePrimaryKey := "`" + strings.ReplaceAll(primaryKey, "`", "``") + "`"
	rows, err := e.db.Query("SELECT * FROM "+safeTableName+" WHERE "+safePrimaryKey+" = ? LIMIT 1", id)
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

func insertTableRecord(e *DbExplorer, table string, data map[string]any) (string, int64, error) {
	safeTableName, fields, err := getSafeTableNameAndFields(e, table)
	if err != nil {
		return "", 0, err
	}

	primaryKeyField, err := findTablePrimaryKeyField(fields)
	if err != nil {
		return "", 0, err
	}
	

	columns, values, err := utils.GetDatabaseFieldsAndValues(data, fields, true)
	if err != nil {
		return "", 0, err
	}

	query := ""
	if len(columns) == 0 {
		query = "INSERT INTO " + safeTableName + " () VALUES ()"
	} else {
		safeColumns := make([]string, 0, len(columns))
		for _, col := range columns {
			safeColumns = append(safeColumns, "`"+strings.ReplaceAll(col, "`", "``")+"`")
		}
		query = "INSERT INTO " + safeTableName + " (" + strings.Join(safeColumns, ", ") + ") VALUES (" + strings.Join(values, ", ") + ")"
	}

	result, err := e.db.Exec(query)
	if err != nil {
		return "", 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", 0, err
	}

	return primaryKeyField.Name, id, nil
}

func updateTableRecord(e *DbExplorer, table string, id int64, data map[string]any) error {
	safeTableName, fields, err := getSafeTableNameAndFields(e, table)
	if err != nil {
		return err
	}

	for _, field := range fields {
		if field.IsPrimary {
			if _, exists := data[field.Name]; exists {
				return fmt.Errorf("field %s have invalid type", field.Name)
			}
		}
	}

	columns, values, err := utils.GetDatabaseFieldsAndValues(data, fields, false)
	if err != nil {
		return err
	}

	if len(columns) == 0 {
		return nil
	}

	primaryKey, err := utils.GetPrimaryKeyColumn(e.db, safeTableName)
	if err != nil {
		return err
	}

	safePrimaryKey := "`" + strings.ReplaceAll(primaryKey, "`", "``") + "`"
	setParts := make([]string, 0, len(columns))
	for i, col := range columns {
		safeCol := "`" + strings.ReplaceAll(col, "`", "``") + "`"
		setParts = append(setParts, safeCol+" = "+values[i])
	}

	query := "UPDATE " + safeTableName + " SET " + strings.Join(setParts, ", ") + " WHERE " + safePrimaryKey + " = ?"

	_, err = e.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func getSafeTableNameAndFields(e *DbExplorer, table string) (string, []utils.TableFieldInfo, error) {
	safeTableName, err := utils.GetSafeTableName(e.tables, table)
	if err != nil {
		return "", nil, err
	}

	rows, err := e.db.Query("SHOW FULL COLUMNS FROM " + safeTableName)
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()

	fields := make([]utils.TableFieldInfo, 0)
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
			return "", nil, err
		}

		required := strings.ToUpper(null) == "NO" && !strings.Contains(strings.ToLower(extra), "auto_increment") && !def.Valid

		fields = append(fields, utils.TableFieldInfo{
			Name:      field,
			TypeName:  typeName,
			Required:  required,
			IsPrimary: key == "PRI",
		})
	}

	if err := rows.Err(); err != nil {
		return "", nil, err
	}

	return safeTableName, fields, nil
}
