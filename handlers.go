package main

import (
	"encoding/json"
	"net/http"

	"db_explorer/utils"
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
		limit, offset, err := utils.GetLimitOffset(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		records, err := getTableRecords(e, table, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeResponse(w, http.StatusOK, map[string]interface{}{
			"records": records,
		})
	case http.MethodPut:
		data := make(map[string]any)
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			writeError(w, http.StatusBadRequest, "bad request")
			return
		}

		id, err := insertTableRecord(e, table, data)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeResponse(w, http.StatusOK, map[string]any{
			"id": id,
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "bad method")
	}
}

func (e *DbExplorer) handleTableRecord(w http.ResponseWriter, r *http.Request, table string, id int64) {

	switch r.Method {
	case http.MethodGet:
		// TODO: get record by id
		record, err := getTableRecord(e, table, id)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeResponse(w, http.StatusOK, map[string]interface{}{
			"record": record,
		})
	case http.MethodPost:
		// TODO: update record by id
		writeError(w, http.StatusNotImplemented, "method POST for record is not implemented yet")
	case http.MethodDelete:
		deleted, err := deleteTableRecord(e, table, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeResponse(w, http.StatusOK, map[string]any{
			"deleted": deleted,
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "bad method")
	}
}
