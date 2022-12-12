package utils

import (
	"strings"
)

// NormalizeGroup normalizes group name
func NormalizeGroup(title string, path string) string {
	if title != "" {
		return title
	}
	n := strings.Index(path, "{")
	if n != -1 {
		path = path[0 : n-1]
	}
	n = strings.Index(path, ":")
	if n != -1 {
		path = path[0 : n-1]
	}
	if len(path) > 0 {
		path = path[1:]
	}
	group := strings.ReplaceAll(path, "/", "_")
	if group == "" {
		group = "root"
	}
	return group
}
