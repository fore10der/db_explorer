package main

import (
	"net/http"
)

func (e *DbExplorer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "bad method")
		return
	}

	tables := getTables(e)

	writeResponse(w, http.StatusOK, map[string]interface{}{
		"tables": tables,
	})
}

func (e *DbExplorer) handleTableCollection(w http.ResponseWriter, r *http.Request, table string) {
	switch r.Method {
	case http.MethodGet:
		records, err := getTableRecords(e, table)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeResponse(w, http.StatusOK, records)
	case http.MethodPut:
		// TODO: create record in table
		writeError(w, http.StatusNotImplemented, "method PUT for table is not implemented yet")
	default:
		writeError(w, http.StatusMethodNotAllowed, "bad method")
	}
}

func (e *DbExplorer) handleTableRecord(w http.ResponseWriter, r *http.Request, table string, id int64) {
	_ = table
	_ = id

	switch r.Method {
	case http.MethodGet:
		// TODO: get record by id
		writeError(w, http.StatusNotImplemented, "method GET for record is not implemented yet")
	case http.MethodPost:
		// TODO: update record by id
		writeError(w, http.StatusNotImplemented, "method POST for record is not implemented yet")
	case http.MethodDelete:
		// TODO: delete record by id
		writeError(w, http.StatusNotImplemented, "method DELETE for record is not implemented yet")
	default:
		writeError(w, http.StatusMethodNotAllowed, "bad method")
	}
}
