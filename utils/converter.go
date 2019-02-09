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

// SplitPath splits the given full path into the path and the name of the file/dir
func SplitPath(origPath string) (path, name string) {
	origPath = ConvertToSlash(origPath, false)
	if origPath == "/." || origPath == "/" || origPath == "." || origPath == "" {
		return "/", ""
	}

	if strings.HasSuffix(origPath, "/") {
		origPath = origPath[:len(origPath)-1]
	}

	path = ConvertToSlash(filepath.Dir(origPath), true)
	if strings.HasSuffix(path, "./") {
		path = path[:len(path)-2]
	}

	name = filepath.Base(origPath)
	return
}
