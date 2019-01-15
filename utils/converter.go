package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var multiSlashReg = regexp.MustCompile(`(\/)+`)

// ConvertToSlash takes a string (most likely a path) and converts all back-slashes to normal slashes
func ConvertToSlash(path string, isPath bool) string {
	path = filepath.Clean(path)
	path = strings.Replace(path, "\\", "/", -1)
	path = multiSlashReg.ReplaceAllString(path, "/")
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	if isPath && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return path
}

// ConvertToCleanEmail removes all leading & trailing whitespaces and makes it lowercase
func ConvertToCleanEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}
