package repository

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

func TestFileInfoRepository(t *testing.T) {
	fileOrig0 := &models.FileInfo{OwnerID: 1, ParentID: 0, Path: "/", Name: "fileOrig0"}
	fileOrig1 := &models.FileInfo{OwnerID: 2, ParentID: 0, Path: "/", Name: "fileOrig1"}
	dbName := "fileInfoTest.db"

	cleanDBFiles := func() {
		os.Remove(dbName)
	}

	cleanDBFiles()
	//defer cleanDBFiles()

	var rep *FileInfoRepository

	success := t.Run("create connection and repository", func(t *testing.T) {
		err := InitDatabaseConnection("", "", "", "", 0, dbName)
		if err != nil {
			t.Fatalf("Failed to connect to gorm database: %v", err)
		}

		rep, err = CreateFileInfoRepository()
		if err != nil {
			t.Fatalf("Failed to create user repository: %v", err)
		}
	})
	if !success {
		t.Skip("Further test skipped due to setup failing")
	}

	t.Run("empty repository", func(t *testing.T) {
		count, err := rep.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count > 0 {
			t.Errorf("Count greater than zero for empty file info repository: %d", count)
		}
	})

	success = t.Run("create file infos", func(t *testing.T) {
		err := rep.Create(fileOrig0)
		if err != nil {
			t.Errorf("Failed to create fileOrig0: %v", err)
		}
		err = rep.Create(fileOrig1)
		if err != nil {
			t.Errorf("Failed to create fileOrig1: %v", err)
		}
	})
	if !success {
		t.Skip("Further tests skipped due to no created file infos")
	}

	var shareRep *ShareEntryRepository
	shareEntry0 := &models.ShareEntry{FileID: fileOrig0.ID}
	shareEntry1 := &models.ShareEntry{FileID: fileOrig1.ID}
	shareEntry2 := &models.ShareEntry{FileID: fileOrig1.ID}

	success = t.Run("create share entry repository and created needed share entries", func(t *testing.T) {
		var err error
		shareRep, err = CreateShareEntryRepository()
		if err != nil {
			t.Fatalf("Failed to initialize share entry repository: %v", err)
		}

		err = shareRep.Create(shareEntry0)
		if err != nil {
			t.Errorf("Failed to create shareEntry0: %v", err)
		}
		err = shareRep.Create(shareEntry1)
		if err != nil {
			t.Errorf("Failed to create shareEntry1: %v", err)
		}
		err = shareRep.Create(shareEntry2)
		if err != nil {
			t.Errorf("Failed to create shareEntry2: %v", err)
		}
	})
	if !success {
		t.Skip("Skipping further tests due to no created share entries")
	}

	fileShared0 := &models.FileInfo{OwnerID: 2, ShareID: shareEntry0.ID, ParentID: 0, Path: "/", Name: "fileShared0"}
	fileShared1 := &models.FileInfo{OwnerID: 1, ShareID: shareEntry1.ID, ParentID: 0, Path: "/", Name: "fileShared1"}
	fileShared2 := &models.FileInfo{OwnerID: 3, ShareID: shareEntry2.ID, ParentID: 0, Path: "/", Name: "fileShared2"}

	success = t.Run("create shared files", func(t *testing.T) {
		err := rep.Create(fileShared0)
		if err != nil {
			t.Errorf("Failed to create fileShared0: %v", err)
		}
		err = rep.Create(fileShared1)
		if err != nil {
			t.Errorf("Failed to create fileShared1: %v", err)
		}
		err = rep.Create(fileShared2)
		if err != nil {
			t.Errorf("Failed to create fileShared2: %v", err)
		}
	})
	if !success {
		t.Skip("Skipping further tests due to no created shared files")
	}

	var starRep *StarRepository
	star0 := &models.Star{FileID: fileOrig0.ID, UserID: fileOrig0.OwnerID}
	star1 := &models.Star{FileID: shareEntry0.FileID, UserID: fileShared0.OwnerID}
	star2 := &models.Star{FileID: fileOrig1.ID, UserID: fileOrig1.OwnerID}

	success = t.Run("create star repository and create stars", func(t *testing.T) {
		var err error
		starRep, err = CreateStarRepository()
		if err != nil {
			t.Fatalf("Failed to create star repository: %v", err)
		}

		err = starRep.Create(star0)
		if err != nil {
			t.Errorf("Failed to create star0: %v", err)
		}
		err = starRep.Create(star1)
		if err != nil {
			t.Errorf("Failed to create star1: %v", err)
		}
		err = starRep.Create(star2)
		if err != nil {
			t.Errorf("Failed to create star2: %v", err)
		}
	})
	if !success {
		t.Skip("Skipping further tests due to no stars")
	}

	t.Run("correct repository count", func(t *testing.T) {
		count, err := rep.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 5 {
			t.Errorf("Count unqual to five for filled file info repository: %d", count)
		}
	})

	t.Run("get file info", func(t *testing.T) {
		readBackFileInfo, err := rep.GetByID(fileOrig0.ID)
		if err != nil {
			t.Errorf("Failed to get file info orig 0 by ID: %v", err)
		}
		if !reflect.DeepEqual(readBackFileInfo, fileOrig0) {
			t.Error("Read back orig file 0 by ID and orig file 0 not deeply equal")
		}
		readBackFileInfo, err = rep.GetByPath(fileOrig1.OwnerID, fileOrig1.Path, fileOrig1.Name)
		if err != nil {
			t.Errorf("Failed to get file info orig 1 by path: %v", err)
		}
		if !readBackFileInfo.Starred {
			t.Error("Read back orig file 1 is not starred")
		}
		readBackFileInfo.Starred = false
		if !reflect.DeepEqual(readBackFileInfo, fileOrig1) {
			t.Error("Read back orig file 1 by path and orig file 1 not deeply equal")
		}
	})
}
