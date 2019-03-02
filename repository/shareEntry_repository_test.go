package repository

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

var testShareEntrySetupFailed = false
var testShareEntryDBName = "shareEntryTest.db"
var testShareEntry0 = &models.ShareEntry{}
var testShareEntry1 = &models.ShareEntry{}
var testShareEntry2 = &models.ShareEntry{}
var testShareEntryFileOrig0 = &models.FileInfo{OwnerID: 1}
var testShareEntryFileOrig1 = &models.FileInfo{OwnerID: 2}
var testShareEntryFileShared0 = &models.FileInfo{OwnerID: 2}
var testShareEntryFileShared1 = &models.FileInfo{OwnerID: 1}
var testShareEntryFileShared2 = &models.FileInfo{OwnerID: 3}

func testShareEntryCleanup() {
	os.Remove(testShareEntryDBName)
}

func testShareEntrySetup() *ShareEntryRepository {
	testShareEntryCleanup()
	InitSQLDatabaseConnection("", "", "", "", 0, testShareEntryDBName)
	rep, _ := CreateShareEntryRepository()
	return rep
}

func testShareEntryInsert(rep *ShareEntryRepository) {
	testShareEntry0.FileID = testShareEntryFileOrig0.ID
	testShareEntry1.FileID = testShareEntryFileOrig1.ID
	testShareEntry2.FileID = testShareEntryFileOrig1.ID
	rep.Create(testShareEntry0)
	rep.Create(testShareEntry1)
	rep.Create(testShareEntry2)
}

func testShareEntryInsertComplete(rep *ShareEntryRepository) {
	fileRep, _ := CreateFileInfoRepository()
	fileRep.Create(testShareEntryFileOrig0)
	fileRep.Create(testShareEntryFileOrig1)

	testShareEntryInsert(rep)

	testShareEntryFileShared0.ShareID = testShareEntry0.ID
	testShareEntryFileShared1.ShareID = testShareEntry1.ID
	testShareEntryFileShared2.ShareID = testShareEntry2.ID
	fileRep.Create(testShareEntryFileShared0)
	fileRep.Create(testShareEntryFileShared1)
	fileRep.Create(testShareEntryFileShared2)
}

func TestCreateShareEntryRepository(t *testing.T) {
	testShareEntryCleanup()
	defer testShareEntryCleanup()

	err := InitSQLDatabaseConnection("", "", "", "", 0, testShareEntryDBName)
	if err != nil {
		t.Errorf("Failed to connect to gorm database: %v", err)
	}

	_, err = CreateShareEntryRepository()
	if err != nil {
		t.Errorf("Failed to create share entry repository: %v", err)
	}

	if t.Failed() {
		testShareEntrySetupFailed = true
	}
}

func TestCreateShareEntry(t *testing.T) {
	if testShareEntrySetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testShareEntryCleanup()
	rep := testShareEntrySetup()

	err := rep.Create(testShareEntry0)
	if err != nil {
		t.Errorf("Failed to create shareEntry0: %v", err)
	}

	if t.Failed() {
		testShareEntrySetupFailed = true
	}
}

func TestCountShareEntry(t *testing.T) {
	if testShareEntrySetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testShareEntryCleanup()
	rep := testShareEntrySetup()

	count, err := rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count > 0 {
		t.Errorf("Count greater than zero for empty share entry repository: %d", count)
	}

	testShareEntryInsert(rep)

	count, err = rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 3 {
		t.Errorf("Count unequal to zero for for filled share entry repository: %d", count)
	}
}

func TestShareEntryGetByID(t *testing.T) {
	if testShareEntrySetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testShareEntryCleanup()
	rep := testShareEntrySetup()

	testShareEntryInsertComplete(rep)

	readBackShareEntry, err := rep.GetByID(testShareEntry0.ID)
	if err != nil {
		t.Errorf("Failed to read back testShareEntry0 by ID: %v", err)
	}
	expRes := &models.ShareEntry{ID: testShareEntry0.ID, FileID: testShareEntry0.FileID, OwnerID: testShareEntryFileOrig0.OwnerID, SharedWithID: testShareEntryFileShared0.OwnerID}
	if !reflect.DeepEqual(readBackShareEntry, expRes) {
		t.Errorf("Read back testShareEntry0 and expected result for testShareEntry0 not deeply equal: %v != %v", readBackShareEntry, expRes)
	}
}

func TestShareEntryGetByFileID(t *testing.T) {
	if testShareEntrySetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testShareEntryCleanup()
	rep := testShareEntrySetup()

	testShareEntryInsertComplete(rep)

	readBackShareEntries, err := rep.GetByFileID(testShareEntry1.FileID)
	if err != nil {
		t.Errorf("Failed to read back share entries with file id '%d': %v", testShareEntry1.FileID, err)
	}
	if len(readBackShareEntries) != 2 {
		t.Errorf("Length of read back share entries with file id '%d' is unequal to 2: %d", testShareEntry1.FileID, len(readBackShareEntries))
	}
}

func TestShareEntryGetByIDForUser(t *testing.T) {
	if testShareEntrySetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testShareEntryCleanup()
	rep := testShareEntrySetup()

	testShareEntryInsertComplete(rep)

	readBackShareEntry, err := rep.GetByIDForUser(testShareEntry1.ID, testShareEntryFileOrig1.OwnerID)
	if err != nil {
		t.Errorf("Failed to read back shareEntry1 by ID and owner id of testShareEntryFileOrig1: %v", err)
	}
	expRes := &models.ShareEntry{ID: testShareEntry1.ID, FileID: testShareEntry1.FileID, OwnerID: testShareEntryFileOrig1.OwnerID, SharedWithID: testShareEntryFileShared1.OwnerID}
	if !reflect.DeepEqual(readBackShareEntry, expRes) {
		t.Errorf("Read back sharedEntry1 and expected result for testShareEntry1 not deeply equal: %v != %v", readBackShareEntry, expRes)
	}

	_, err = rep.GetByIDForUser(testShareEntry0.ID, 9999)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Succeeded to read share entry for wrong user or error is not 'record not found': %v", err)
	}
}

func TestShareEntryDelete(t *testing.T) {
	if testShareEntrySetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testShareEntryCleanup()
	rep := testShareEntrySetup()

	testShareEntryInsertComplete(rep)

	err := rep.Delete(testShareEntry2.ID)
	if err != nil {
		t.Fatalf("Failed to delete testShareEntry2: %v", err)
	}

	_, err = rep.GetByID(testShareEntry2.ID)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Succeeded to read deleted share entry or error is not 'record not found': %v", err)
	}
	readBackShareEntries, err := rep.GetByFileID(testShareEntry2.FileID)
	if err != nil {
		t.Errorf("Failed to get share entries after deletion with file id '%d': %v", testShareEntry2.FileID, err)
	}
	if len(readBackShareEntries) != 1 {
		t.Errorf("Length of read back share entries after deletion with file id '%d' is unequal to 2: %d", testShareEntry2.FileID, len(readBackShareEntries))
	}
	expRes := &models.ShareEntry{ID: testShareEntry1.ID, FileID: testShareEntry1.FileID, OwnerID: testShareEntryFileOrig1.OwnerID, SharedWithID: testShareEntryFileShared1.OwnerID}
	if !reflect.DeepEqual(readBackShareEntries[0], expRes) {
		t.Errorf("Remaining share entry for fileID '%d' it not deeply equal to expected result of not deleted testShareEntry1", testShareEntry2.FileID)
	}

	count, err := rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count after share entry deletion: %v", err)
	}
	if count != 2 {
		t.Errorf("Count unequal to two after share entry deletion: %d", count)
	}
}
