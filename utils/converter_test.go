package utils

import "testing"

func TestConvertToSlash(t *testing.T) {
	type InputStruct struct {
		Path  string
		IsDir bool
	}
	var l = map[InputStruct]string{
		{"/path/to/file", false}:      "/path/to/file",
		{"path/to/file", false}:       "/path/to/file",
		{"\\path/to/file", false}:     "/path/to/file",
		{"\\path\\to\\file", false}:   "/path/to/file",
		{"\\path\\\\to\\file", false}: "/path/to/file",
		{"/path/to/dir/", true}:       "/path/to/dir/",
		{"path/to/dir/", true}:        "/path/to/dir/",
		{"\\path/to/dir\\", true}:     "/path/to/dir/",
		{"\\path\\to\\dir/", true}:    "/path/to/dir/",
		{"\\path\\\\to\\dir\\", true}: "/path/to/dir/",
		{"dir/", true}:                "/dir/",
		{"dir\\", true}:               "/dir/",
		{"/dir", true}:                "/dir/",
		{"\\dir/", true}:              "/dir/",
		{"/dir\\", true}:              "/dir/",
		{"dir", true}:                 "/dir/",
	}

	for input, expOutput := range l {
		if output := ConvertToSlash(input.Path, input.IsDir); output != expOutput {
			t.Errorf("Expected result '%s' for input '%v' but got: '%s'", expOutput, input, output)
		}
	}
}

func TestConvertToCleanEmail(t *testing.T) {
	var l = map[string]string{
		"UpperCase@Email.Com":                  "uppercase@email.com",
		"    leadingspaces@email.com":          "leadingspaces@email.com",
		"trailingspaces@email.com       ":      "trailingspaces@email.com",
		"   \t leadingtrailing@email.com   \t": "leadingtrailing@email.com",
		"\t  CombIned@EmaiL.cOm  \t  ":         "combined@email.com",
		"correct@email.com":                    "correct@email.com",
	}

	for input, expOutput := range l {
		if output := ConvertToCleanEmail(input); output != expOutput {
			t.Errorf("Expected result '%s' for input '%s' but got: '%s'", expOutput, input, output)
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

	for input, expOutput := range l {
		if path, name := SplitPath(input); path != expOutput[0] || name != expOutput[1] {
			t.Errorf("Expected result %v for input %s but got: %v and %v", expOutput, input, path, name)
		}
	}
}
