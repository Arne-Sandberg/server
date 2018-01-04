package utils

import "testing"

func TestRandomString(t *testing.T) {
	var l = []int{1, 5, 10, 20}
	for _, v := range l {
		if length := len(RandomString(v)); length != v {
			t.Errorf("Expected string of length %d, but got %d", v, length)
		}
	}

	if RandomString(10) == RandomString(10) {
		t.Error("Expected two different random strings but got two times the same")
	}
}
