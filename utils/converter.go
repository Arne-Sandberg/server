package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var multiSlashReg = regexp.MustCompile(`(\/)+`)

// ConvertToSlash takes a string (most likely a path) and converts all back-slashes to normal slashes
func ConvertToSlash(path string, isDir bool) string {
	path = filepath.Clean(path)
	path = strings.Replace(path, "\\", "/", -1)
	path = multiSlashReg.ReplaceAllString(path, "/")
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	if isDir && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return path
}
