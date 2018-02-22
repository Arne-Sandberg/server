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
			t.Errorf("Expected result %s for input %s but got: %s", expOutput, input, output)
		}
	}
}
