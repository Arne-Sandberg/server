package fs

import (
	"testing"
)

func TestRejectInsanePath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"simple valid dir", args{"/this/should/work"}, nil},
		{"complex valid dir", args{"/this/should/.also/Work!/"}, nil},
		{"invalid path upward between", args{"/this/../should/fail"}, ErrUpwardsNavigation},
		{"invalid path upward start", args{"../not/valid!"}, ErrUpwardsNavigation},
		{"invalid path upward end", args{"/invalid/path!/.."}, ErrUpwardsNavigation},
		{"invalid path <", args{"/path/</to/failure"}, ErrForbiddenPathName},
		{"invalid path >", args{"/path/>/to/the/right"}, ErrForbiddenPathName},
		{"invalid path :", args{"/typing/:/paths/is/boring"}, ErrForbiddenPathName},
		{"invalid path \"", args{"/quoted/\"/path"}, ErrForbiddenPathName},
		{"invalid path \\", args{"/those/\\/Windows/users/"}, ErrForbiddenPathName},
		{"invalid path |", args{"/repetitive/|/task!"}, ErrForbiddenPathName},
		{"invalid path ?", args{"/are/we/there/yet?"}, ErrForbiddenPathName},
		{"invalid path *", args{"/finally/*_*"}, ErrForbiddenPathName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DiskFilesystem{"."}
			if err := dfs.rejectInsanePath(tt.args.path); err != tt.wantErr {
				t.Errorf("rejectIfNavigatingUpwards() in test %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
