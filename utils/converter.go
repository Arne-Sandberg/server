package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var multiSlashReg = regexp.MustCompile(`(\/)+`)

// ConvertToSlash takes a string (most likely a path) and converts all back-slashes to normal slashes
func ConvertToSlash(path string) string {
	path = filepath.Clean(path)
	path = strings.Replace(path, "\\", "/", -1)
	return multiSlashReg.ReplaceAllString(path, "/")
}