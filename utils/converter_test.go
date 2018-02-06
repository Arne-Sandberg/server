package utils

import "testing"

func TestConvertToSlash(t *testing.T) {
	var l = map[string]string{
		"/path/to/file":      "/path/to/file",
		"\\path/to/file":     "/path/to/file",
		"\\path\\to\\file":   "/path/to/file",
		"\\path\\\\to\\file": "/path/to/file",
	}

	for input, expOutput := range l {
		if output := ConvertToSlash(input); output != expOutput {
			t.Errorf("Expected result %s for input %s but got: %s", expOutput, input, output)
		}
	}
}
