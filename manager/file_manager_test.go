package manager

import (
	"testing"

	"github.com/freecloudio/server/models"
)

func TestGetUserPath(t *testing.T) {
	var l = map[int64]string{
		0: "/0",
		1: "/1",
	}

	mgr := FileManager{}
	for input, expOutput := range l {
		if output := mgr.getUserPath(&models.User{ID: input}); output != expOutput {
			t.Errorf("Expected result %s for input %v but got: %s", expOutput, input, output)
		}
	}
}

func TestSplitPath(t *testing.T) {
	var l = map[string][2]string{
		"/hello/dear/file.txt": {"/hello/dear/", "file.txt"},
		"/file2.txt":           {"/", "file2.txt"},
		"/":                    {"/", ""},
		"\\":                   {"/", ""},
		".":                    {"/", ""},
		"\\testFolder\\":       {"/", "testFolder"},
		"/testFolder":          {"/", "testFolder"},
		"/.tmp":                {"/", ".tmp"},
		"testFolder/":          {"/", "testFolder"},
		".tmp/":                {"/", ".tmp"},
		"/testFolder/":         {"/", "testFolder"},
		"/.tmp/":               {"/", ".tmp"},
		"testFolder":           {"/", "testFolder"},
		".tmp":                 {"/", ".tmp"},
	}

	mgr := FileManager{}
	for input, expOutput := range l {
		if path, name := mgr.splitPath(input); path != expOutput[0] || name != expOutput[1] {
			t.Errorf("Expected result %v for input %s but got: %v and %v", expOutput, input, path, name)
		}
	}
}
