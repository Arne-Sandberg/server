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


