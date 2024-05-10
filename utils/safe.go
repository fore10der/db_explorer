package utils

import (
	"fmt"
	"strings"
)

func GetSafeTableName(tables map[string]struct{}, table string) (string, error) {
	if _, ok := tables[table]; !ok {
		return "", fmt.Errorf("unknown table")
	}

	return "`" + strings.ReplaceAll(table, "`", "``") + "`", nil
}