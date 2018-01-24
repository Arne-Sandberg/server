package fs

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/freecloud/models"
)

func TestDiskFilesystem_rejectIfNavigatingUpwards(t *testing.T) {
	type fields struct {
		base string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"simple valid dir" },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dfs := &DiskFilesystem{
				base: ".",
			}
			if err := dfs.rejectIfNavigatingUpwards(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("DiskFilesystem.rejectIfNavigatingUpwards() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
