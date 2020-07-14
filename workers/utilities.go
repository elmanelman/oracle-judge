package workers

import "strings"

func prepareSelectionSolution(query string) string {
	query = strings.TrimSpace(query)
	query = strings.ReplaceAll(query, ";", "")
	query = strings.ReplaceAll(query, "\n", " ")
	return query
}
