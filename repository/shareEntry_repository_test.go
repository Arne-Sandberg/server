package repository

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

func TestShareEntryRepository(t *testing.T) {
	dbName := "shareEntryTest.db"

	cleanDBFiles := func() {
		os.Remove(dbName)
	}

	cleanDBFiles()
	defer cleanDBFiles()

	var rep *ShareEntryRepository

	success := t.Run("create connection and repository", func(t *testing.T) {
		err := InitDatabaseConnection("", "", "", "", 0, dbName)
		if err != nil {
			t.Fatalf("Failed to connect to gorm database: %v", err)
		}

		rep, err = CreateShareEntryRepository()
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
			t.Errorf("Count greater than zero for empty share entry repository: %d", count)
		}
	})

	var fileRep *FileInfoRepository
	fileOrig0 := &models.FileInfo{OwnerID: 1}
	fileOrig1 := &models.FileInfo{OwnerID: 2}

	success = t.Run("create file info repository and create needed files", func(t *testing.T) {
		var err error
		fileRep, err = CreateFileInfoRepository()
		if err != nil {
			t.Fatalf("Failed to initialize file info repository: %v", err)
		}

		err = fileRep.Create(fileOrig0)
		if err != nil {
			t.Errorf("Failed to create fileOrig0: %v", err)
		}
		err = fileRep.Create(fileOrig1)
		if err != nil {
			t.Errorf("Failed to create fileOrig1: %v", err)
		}
	})
	if !success {
		t.Skip("Further test skipped du to no created original files")
	}

	shareEntry0 := &models.ShareEntry{FileID: fileOrig0.ID}
	shareEntry1 := &models.ShareEntry{FileID: fileOrig1.ID}
	shareEntry2 := &models.ShareEntry{FileID: fileOrig1.ID}

	success = t.Run("create share entries", func(t *testing.T) {
		err := rep.Create(shareEntry0)
		if err != nil {
			t.Errorf("Failed to create shareEntry0: %v", err)
		}
		err = rep.Create(shareEntry1)
		if err != nil {
			t.Errorf("Failed to create shareEntry1: %v", err)
		}
		err = rep.Create(shareEntry2)
		if err != nil {
			t.Errorf("Failed to create shareEntry2: %v", err)
		}
	})
	if !success {
		t.Skip("Skipping further tests due to no created share entries")
	}

	fileShared0 := &models.FileInfo{OwnerID: 2, ShareID: shareEntry0.ID}
	fileShared1 := &models.FileInfo{OwnerID: 1, ShareID: shareEntry1.ID}
	fileShared2 := &models.FileInfo{OwnerID: 3, ShareID: shareEntry2.ID}

	success = t.Run("create shared files", func(t *testing.T) {
		err := fileRep.Create(fileShared0)
		if err != nil {
			t.Errorf("Failed to create fileShared0: %v", err)
		}
		err = fileRep.Create(fileShared1)
		if err != nil {
			t.Errorf("Failed to create fileShared1: %v", err)
		}
		err = fileRep.Create(fileShared2)
		if err != nil {
			t.Errorf("Failed to create fileShared2: %v", err)
		}
	})
	if !success {
		t.Skip("Skipping further tests due to no created shared files")
	}

	t.Run("correct count after creating share entries", func(t *testing.T) {
		count, err := rep.Count()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 3 {
			t.Errorf("Count unequal to zero for for filled share entry repository: %d", count)
		}
	})

	t.Run("correct read back of created share entries", func(t *testing.T) {
		readBackShareEntry, err := rep.GetByID(shareEntry0.ID)
		if err != nil {
			t.Errorf("Failed to read back shareEntry0 by ID: %v", err)
		}
		expRes := &models.ShareEntry{ID: shareEntry0.ID, FileID: shareEntry0.FileID, OwnerID: fileOrig0.OwnerID, SharedWithID: fileShared0.OwnerID}
		if !reflect.DeepEqual(readBackShareEntry, expRes) {
			t.Error("Read back shareEntry0 and expected result for shareEntry0 not deeply equal")
		}
		readBackShareEntries, err := rep.GetByFileID(shareEntry1.FileID)
		if err != nil {
			t.Errorf("Failed to read back share entries with file id '%d': %v", shareEntry1.FileID, err)
		}
		if len(readBackShareEntries) != 2 {
			t.Errorf("Length of read back share entries with file id '%d' is unequal to 2: %d", shareEntry1.FileID, len(readBackShareEntries))
		}
	})

	delSuccess := t.Run("delete share entry", func(t *testing.T) {
		err := rep.Delete(shareEntry2.ID)
		if err != nil {
			t.Errorf("Failed to delete shareEntry2: %v", err)
		}
	})

	if delSuccess {
		t.Run("correct read back after deleting share entry", func(t *testing.T) {
			_, err := rep.GetByID(shareEntry2.ID)
			if err == nil || !IsRecordNotFoundError(err) {
				t.Errorf("Succeeded to read deleted share entry or error is not 'record not found': %v", err)
			}
			readBackShareEntries, err := rep.GetByFileID(shareEntry2.FileID)
			if err != nil {
				t.Fatalf("Failed to get share entries after deletion with file id '%d': %v", shareEntry2.FileID, err)
			}
			if len(readBackShareEntries) != 1 {
				t.Fatalf("Length of read back share entries after deletion with file id '%d' is unequal to 2: %d", shareEntry2.FileID, len(readBackShareEntries))
			}
			expRes := &models.ShareEntry{ID: shareEntry1.ID, FileID: shareEntry1.FileID, OwnerID: fileOrig1.OwnerID, SharedWithID: fileShared1.OwnerID}
			if !reflect.DeepEqual(readBackShareEntries[0], expRes) {
				t.Errorf("Remaining share entry for fileID '%d' it not deeply equal to expected result of not deleted shareEntry1", shareEntry2.FileID)
			}
		})
	}

	if delSuccess {
		t.Run("correct count after share entry deletion", func(t *testing.T) {
			count, err := rep.Count()
			if err != nil {
				t.Fatalf("Failed to get count after share entry deletion: %v", err)
			}
			if count != 2 {
				t.Errorf("Count unequal to two after share entry deletion: %d", count)
			}
		})
	}
}
