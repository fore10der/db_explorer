package utils

import (
	"net/http"
	"strconv"
)

func GetLimitOffset(r *http.Request) (int, int, error) {
	limit := 5
	offset := 0

	if r.URL.Query().Get("limit") != "" {
		parsedLimit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err == nil {
			limit = parsedLimit
		}
	}

	if r.URL.Query().Get("offset") != "" {
		parsedOffset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err == nil {
			offset = parsedOffset
		}
	}

	return limit, offset, nil
}
