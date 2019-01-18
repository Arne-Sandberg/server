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
