package repository

import (
	"os"
	"testing"
)

func TestFileSystemRepository(t *testing.T) {
	dirName := "testData"
	tmpName := ".tmp"

	clearData := func() {
		os.RemoveAll(dirName)
	}

	clearData()
	//defer clearData()

	var rep *FileSystemRepository

	success := t.Run("create repository", func(t *testing.T) {
		var err error
		rep, err = CreateFileSystemRepository(dirName, tmpName, 1, 0)
		if err != nil {
			t.Errorf("Failed to create fileSystemRepository> %v", err)
		}
	})
	if !success {
		t.Skip("Skip further tests due to failing setup")
	}

	success = t.Run("create dir", func(t *testing.T) {
		created, err := rep.CreateDirectory("1/.tmp")
		if err != nil {
			t.Errorf("Failed to create directory '1/.tmp': %v", err)
		}
		if !created {
			t.Error("Directory '1/.tmp' not created")
		}
		created, err = rep.CreateDirectory("/2")
		if err != nil {
			t.Errorf("Failed to create directory '/2': %v", err)
		}
		if !created {
			t.Error("Directory '/2' not created")
		}
		created, err = rep.CreateDirectory("~/badDir")
		if err != ErrForbiddenPathName {
			t.Errorf("Error for forbidden file name is unequal to ErrForbiddenFileName: %v", err)
		}
		if created {
			t.Error("Directory with forbidden name created")
		}
	})

	success = t.Run("new file handle", func(t *testing.T) {
		file, err := rep.CreateHandle("1/.tmp/testfile.txt")
		if err != nil {
			t.Errorf("Failed to create new file handle for '1/testfile.txt': %v", err)
		}
		file.Close()
		_, err = rep.CreateHandle("~/badFile.txt")
		if err != ErrForbiddenPathName {
			t.Errorf("Error for forbidden file name is unequal to ErrForbiddenPathName: %v", err)
		}
	})
	if !success {
		t.Skip("Skip further tests due to failing setup")
	}

	t.Run("close repository", func(t *testing.T) {
		err := rep.Close()
		if err != nil {
			t.Errorf("Failed to close repository: %v", err)
		}
	})
}
