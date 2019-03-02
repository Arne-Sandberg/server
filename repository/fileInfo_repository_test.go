package repository

import (
	"os"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

var testFileInfoSetupFailed = false
var testFileInfoDBName = "fileInfoTest.db"
var testFileInfoOrig0 = &models.FileInfo{OwnerID: 1, ParentID: 101, Path: "/", Name: "fileOrig0"}
var testFileInfoOrig1 = &models.FileInfo{OwnerID: 2, ParentID: 102, Path: "/", Name: "fileOrig1"}
var testFileInfoShareEntry0 = &models.ShareEntry{}
var testFileInfoShareEntry1 = &models.ShareEntry{}
var testFileInfoShareEntry2 = &models.ShareEntry{}
var testFileInfoShared0 = &models.FileInfo{OwnerID: 2, ParentID: 102, Path: "/", Name: "fileShared0"}
var testFileInfoShared1 = &models.FileInfo{OwnerID: 1, ParentID: 101, Path: "/", Name: "fileShared1"}
var testFileInfoShared2 = &models.FileInfo{OwnerID: 3, ParentID: 103, Path: "/", Name: "fileShared2"}
var testFileInfoStar0 = &models.Star{}
var testFileInfoStar1 = &models.Star{}
var testFileInfoStar2 = &models.Star{}

func testFileInfoCleanup() {
	os.Remove(testFileInfoDBName)
}

func testFileInfoInsert(rep *FileInfoRepository) {
	rep.Create(testFileInfoOrig0)
	rep.Create(testFileInfoOrig1)
}

func testFileInfoInsertComplete(rep *FileInfoRepository) {
	testFileInfoInsert(rep)
	testFileInfoShareEntry0.FileID = testFileInfoOrig0.ID
	testFileInfoShareEntry1.FileID = testFileInfoOrig1.ID
	testFileInfoShareEntry2.FileID = testFileInfoOrig1.ID
	shareRep, _ := CreateShareEntryRepository()
	shareRep.Create(testFileInfoShareEntry0)
	shareRep.Create(testFileInfoShareEntry1)
	shareRep.Create(testFileInfoShareEntry2)
	testFileInfoShared0.ShareID = testFileInfoShareEntry0.ID
	testFileInfoShared1.ShareID = testFileInfoShareEntry1.ID
	testFileInfoShared2.ShareID = testFileInfoShareEntry2.ID
	rep.Create(testFileInfoShared0)
	rep.Create(testFileInfoShared1)
	rep.Create(testFileInfoShared2)
	testFileInfoStar0.FileID = testFileInfoOrig0.ID
	testFileInfoStar0.UserID = testFileInfoOrig0.OwnerID
	testFileInfoStar1.FileID = testFileInfoShared0.ID
	testFileInfoStar1.UserID = testFileInfoShared0.OwnerID
	testFileInfoStar2.FileID = testFileInfoOrig1.ID
	testFileInfoStar2.UserID = testFileInfoOrig1.OwnerID
	starRep, _ := CreateStarRepository()
	starRep.Create(testFileInfoStar0)
	starRep.Create(testFileInfoStar1)
	starRep.Create(testFileInfoStar2)
}

func testFileInfoSetup() *FileInfoRepository {
	testFileInfoCleanup()
	InitSQLDatabaseConnection("", "", "", "", 0, testFileInfoDBName)
	rep, _ := CreateFileInfoRepository()
	return rep
}

func TestCreateFileInfoRepository(t *testing.T) {
	testFileInfoCleanup()
	defer testFileInfoCleanup()

	err := InitSQLDatabaseConnection("", "", "", "", 0, testFileInfoDBName)
	if err != nil {
		t.Errorf("Failed to connect to gorm database: %v", err)
	}

	_, err = CreateFileInfoRepository()
	if err != nil {
		t.Errorf("Failed to create user repository: %v", err)
	}

	if t.Failed() {
		testFileInfoSetupFailed = true
	}
}

func TestCreateFileInfo(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	err := rep.Create(testFileInfoOrig0)
	if err != nil {
		t.Errorf("Failed to create fileOrig0: %v", err)
	}

	if t.Failed() {
		testFileInfoSetupFailed = true
	}
}

func TestCountFileInfos(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	count, err := rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count > 0 {
		t.Errorf("Count greater than zero for empty file info repository: %d", count)
	}

	testFileInfoInsert(rep)

	count, err = rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 2 {
		t.Errorf("Count unqual to two for filled file info repository: %d", count)
	}
}

func TestFileInfoGetByID(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	readBackFileInfo, err := rep.GetByID(testFileInfoOrig0.ID)
	if err != nil {
		t.Errorf("Failed to get file info orig 0 by ID: %v", err)
	}
	if !reflect.DeepEqual(readBackFileInfo, testFileInfoOrig0) {
		t.Errorf("Read back orig file 0 by ID and orig file 0 not deeply equal: %v != %v", readBackFileInfo, testFileInfoOrig0)
	}
}

func TestFileInfoGetByPath(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	readBackFileInfo, err := rep.GetByPath(testFileInfoOrig1.OwnerID, testFileInfoOrig1.Path, testFileInfoOrig1.Name)
	if err != nil {
		t.Errorf("Failed to get file info orig 1 by path: %v", err)
	}
	if !readBackFileInfo.Starred {
		t.Error("Read back orig file 1 is not starred")
	}
	readBackFileInfo.Starred = false
	if !reflect.DeepEqual(readBackFileInfo, testFileInfoOrig1) {
		t.Errorf("Read back orig file 1 by path and orig file 1 not deeply equal: %v != %v", readBackFileInfo, testFileInfoOrig1)
	}
}

func TestFileInfoGetDirectoryContentByID(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	readBackFileInfos, err := rep.GetDirectoryContentByID(1, 101)
	if err != nil {
		t.Errorf("Failed to get dir content for dir 101 and user 1: %v", err)
	}
	if len(readBackFileInfos) != 2 {
		t.Fatalf("Length of read back dir content of dir 101 and user 1 is unequal to two: %d", len(readBackFileInfos))
	}
	if !readBackFileInfos[0].Starred {
		t.Error("First file of read back dir content of dir 101 and user 1 is not starred")
	}
	readBackFileInfos[0].Starred = false
	if !reflect.DeepEqual(readBackFileInfos[0], testFileInfoOrig0) {
		t.Errorf("First info of read back dir content of dir 101 and user 1 and file orig 0 are not deeply equal: %v != %v", readBackFileInfos[0], testFileInfoOrig0)
	}
	if !reflect.DeepEqual(readBackFileInfos[1], testFileInfoShared1) {
		t.Errorf("Second file of read back dir content of dir 101 and user 1 and file shared 1 are not deeply equal: %v != %v", readBackFileInfos, testFileInfoShared1)
	}
}

func TestFileInfoGetStarredByUser(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	readBackFileInfos, err := rep.GetStarredFileInfosByUser(2)
	if err != nil {
		t.Errorf("Failed to get starred files for user 2: %v", err)
	}
	if len(readBackFileInfos) != 2 {
		t.Error("Length of read back starred file of user 2 is unequal to two")
	}
	if !readBackFileInfos[0].Starred || !readBackFileInfos[1].Starred {
		t.Error("Not all read back starred file infors are starred")
	}
	readBackFileInfos[0].Starred = false
	readBackFileInfos[1].Starred = false
	if !reflect.DeepEqual(readBackFileInfos[0], testFileInfoOrig1) {
		t.Errorf("First file of read back starred files for user 2 and file orig 1 are not deeply equal: %v != %v", readBackFileInfos[0], testFileInfoOrig1)
	}
	if !reflect.DeepEqual(readBackFileInfos[1], testFileInfoShared0) {
		t.Errorf("Second file of read back starred file for user 2 and file shared 0 are not deeply equal: %v != %v", readBackFileInfos[1], testFileInfoShared0)
	}
}

func TestFileInfoSearch(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	searchResults, err := rep.Search(1, "/", "file")
	if err != nil {
		t.Fatalf("Failed to search for '/' and 'file' for user 0: %v", err)
	}
	if len(searchResults) != 2 {
		t.Fatalf("Length of search result unequal to two: %d", len(searchResults))
	}
	if !searchResults[0].Starred {
		t.Error("First result of search not starred")
	}
	searchResults[0].Starred = false
	if !reflect.DeepEqual(searchResults[0], testFileInfoOrig0) {
		t.Errorf("First search result and fileOrig not deeply equal: %v != %v", searchResults[0], testFileInfoOrig0)
	}
	if !reflect.DeepEqual(searchResults[1], testFileInfoShared1) {
		t.Errorf("Second search result and fileShared1 not deeply equal: %v != %v", searchResults[1], testFileInfoShared1)
	}
}

// TODO: Test GetShared, GetSharedWith

func TestDeleteFileInfo(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	err := rep.Delete(testFileInfoShared2.ID)
	if err != nil {
		t.Errorf("Failed to delete fileShared2: %v", err)
	}

	_, err = rep.GetByID(testFileInfoShared2.ID)
	if err == nil || !IsRecordNotFoundError(err) {
		t.Errorf("Reading back deleted file info succeeded or error is not 'record not found': %v", err)
	}
	count, err := rep.Count()
	if err != nil {
		t.Errorf("Failed to get count after deleting file: %v", err)
	}
	if count != 4 {
		t.Errorf("Count after deleting file unequal to 4: %d", count)
	}
}

func TestUpdateFileInfo(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	testFileInfoShared1.Name = "new name"
	err := rep.Update(testFileInfoShared1)
	if err != nil {
		t.Errorf("Failed to update fileShared1: %v", err)
	}
	readBackFileInfo, err := rep.GetByPath(testFileInfoShared1.OwnerID, testFileInfoShared1.Path, testFileInfoShared1.Name)
	if err != nil {
		t.Errorf("Failed to read back updated fileShared1: %v", err)
	}
	readBackFileInfo.Starred = false
	if !reflect.DeepEqual(readBackFileInfo, testFileInfoShared1) {
		t.Errorf("Read back updated fileShared1 and fileShared1 not deeply equal: %v != %v", readBackFileInfo, testFileInfoShared1)
	}
}

func TestDeleteUserFileInfo(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	defer testFileInfoCleanup()
	rep := testFileInfoSetup()

	testFileInfoInsertComplete(rep)

	err := rep.DeleteUserFileInfos(2)
	if err != nil {
		t.Errorf("Failed to delete files for user 2: %v", err)
	}
	count, err := rep.Count()
	if err != nil {
		t.Errorf("Failed to get count after delete user file info: %v", err)
	}
	if count != 3 {
		t.Errorf("Count after deleting user file infos for user 2 is unqual to three: %d", count)
	}
}
