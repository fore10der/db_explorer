package utils

import (
	"fmt"
	"net/http"
	"strconv"
)

func GetLimitOffset(r *http.Request) (int, int, error) {
	limit := 5
	offset := 0

	if r.URL.Query().Get("limit") != "" {
		parsedLimit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid limit")
		}
		limit = parsedLimit
	}

	if r.URL.Query().Get("offset") != "" {
		parsedOffset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid offset")
		}
		offset = parsedOffset
	}

	return limit, offset, nil
}
