package fs

import (
	"testing"

	"github.com/freecloudio/freecloud/models"
)

func TestGetUserPath(t *testing.T) {
	var l = map[models.User]string{
		models.User{ID: 0}: "0",
		models.User{ID: 1}: "1",
	}

	vfs := CreateVirtualFileSystem(nil, nil)
	for input, expOutput := range l {
		if output := vfs.getUserPath(&input); output != expOutput {
			t.Errorf("Expected result %s for input %s but got: %s", expOutput, input, output)
		}
	}
}

func TestSplitPath(t *testing.T) {
	var l = map[string][2]string{
		"/hello/dear/file.txt": [2]string{"/hello/dear/", "file.txt"},
		"/file2.txt":           [2]string{"/", "file2.txt"},
		"/":                    [2]string{"/", ""},
		".":                    [2]string{"/", ""},
	}

	vfs := CreateVirtualFileSystem(nil, nil)
	for input, expOutput := range l {
		if path, name := vfs.splitPath(input); path != expOutput[0] || name != expOutput[1] {
			t.Errorf("Expected result %v for input %s but got: %v and %v", expOutput, input, path, name)
		}
	}
}

func TestRemoveUserFromPath(t *testing.T) {
	var l = map[string]string{
		"/1/path/to/file.txt": "/path/to/file.txt",
		"1/path/to/file.txt":  "/path/to/file.txt",
	}

	vfs := CreateVirtualFileSystem(nil, nil)
	for input, expOutput := range l {
		if path := vfs.removeUserFromPath(&models.User{ID: 1}, input); path != expOutput {
			t.Errorf("Expected result %v for input %s but got: %v", expOutput, input, path)
		}
	}
}
