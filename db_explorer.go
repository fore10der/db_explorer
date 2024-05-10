package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type DbExplorer struct {
	db     *sql.DB
	tables map[string]struct{}
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	explorer := &DbExplorer{
		db:     db,
		tables: make(map[string]struct{}),
	}

	if err := explorer.loadTables(); err != nil {
		return nil, err
	}

	return explorer, nil
}

func (e *DbExplorer) loadTables() error {
	rows, err := e.db.Query("SHOW TABLES")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return err
		}
		e.tables[table] = struct{}{}
	}

	return rows.Err()
}

func (e *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		e.handleRoot(w, r)
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		writeError(w, http.StatusNotFound, "unknown table")
		return
	}

	table := parts[0]
	if _, ok := e.tables[table]; !ok {
		writeError(w, http.StatusNotFound, "unknown table")
		return
	}

	if len(parts) == 1 {
		e.handleTableCollection(w, r, table)
		return
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "record not found")
		return
	}

	e.handleTableRecord(w, r, table, id)
}
