package data

import (
	"strings"
)

func CreateOrderByClause(orderBy []string) string {
	if len(orderBy) == 0 {
		return "ORDER BY id"
	}

	var orderClauses []string
	for _, item := range orderBy {
		if strings.HasPrefix(item, "-") {
			orderClauses = append(orderClauses, strings.TrimPrefix(item, "-")+" DESC")
		} else {
			orderClauses = append(orderClauses, item+" ASC")
		}
	}

	return "ORDER BY " + strings.Join(orderClauses, ", ")
}
