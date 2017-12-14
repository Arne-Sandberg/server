package utils

import "testing"

func TestRandomString(t *testing.T) {
	var l = []int{1, 5, 10, 20}
	for _, v := range l {
		if length := len(RandomString(v)); length != v {
			t.Errorf("Expected string of length %d, but got %d", v, length)
		}
	}
}
