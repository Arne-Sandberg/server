package utils

import (
	"testing"
)

func TestValidatePath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		args   args
		expRes bool
	}{
		{"simple valid dir", args{"/this/should/work"}, true},
		{"complex valid dir", args{"/this/should/.also/Work!/"}, true},
		{"valid path \\", args{"/those/\\/Windows/users/"}, true},
		{"valid short path", args{"home"}, true},
		{"invalid path upward between", args{"/this/../should/fail"}, false},
		{"invalid path upward start", args{"../not/valid!"}, false},
		{"invalid path upward end", args{"/invalid/path!/.."}, false},
		{"invalid path <", args{"/path/</to/failure"}, false},
		{"invalid path >", args{"/path/>/to/the/right"}, false},
		{"invalid path :", args{"/typing/:/paths/is/boring"}, false},
		{"invalid path \"", args{"/quoted/\"/path"}, false},
		{"invalid path |", args{"/repetitive/|/task!"}, false},
		{"invalid path ?", args{"/are/we/there/yet?"}, false},
		{"invalid path *", args{"/finally/*_*"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if res := ValidatePath(tt.args.path); res != tt.expRes {
				t.Errorf("ValidatePath() in test %s result = %v, expRes = %v", tt.name, res, tt.expRes)
			}
		})
	}
}
