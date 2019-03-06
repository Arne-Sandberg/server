package utils

import (
	"testing"
)

func TestValidateUsername(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		args   args
		expRes bool
	}{
		{"only characters", args{"thisisausername"}, true},
		{"only digits", args{"23154222224"}, true},
		{"characters and digits", args{"this1is2a3fancy4name"}, true},
		{"with single dot", args{"my.name"}, true},
		{"with dash", args{"my-name"}, true},
		{"with underscore", args{"my_name"}, true},
		{"forward slash", args{"/adsdad"}, false},
		{"back slash", args{"\\adsdad"}, false},
		{"doubled dot", args{"something..with..double..dots"}, false},
		{"single quote", args{"'myname'"}, false},
		{"double quote", args{"\"myname\""}, false},
		{"greater sign", args{">greatname"}, false},
		{"smaller sign", args{"<smallname"}, false},
		{"smaller sign", args{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if res := ValidateUsername(tt.args.path); res != tt.expRes {
				t.Errorf("ValidateUsername() in test %s result = %v, expRes = %v", tt.name, res, tt.expRes)
			}
		})
	}
}

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
