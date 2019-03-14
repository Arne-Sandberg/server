package repository

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

var testFileInfoSetupFailed = false
var testFileInfoUser0 = &models.User{Email: "user0", Username: "user0"}
var testFileInfoUser0Dir0 = &models.FileInfo{OwnerUsername: testFileInfoUser0.Username, Name: "dir0", Path: "/", IsDir: true}
var testFileInfoUser0Dir0File0 = &models.FileInfo{OwnerUsername: testFileInfoUser0.Username, Name: "file0.txt", Path: "/dir0", IsDir: false}

func testFileInfoSetup() *FileInfoRepository {
	testConnectClearGraph()

	userRep, _ := CreateUserRepository()
	userRep.Create(testFileInfoUser0)

	rep, _ := CreateFileInfoRepository()
	return rep
}

func testFileInfoInsert(rep *FileInfoRepository) {
	rep.CreateUserFolder(testFileInfoUser0.Username)
	rep.Create(testFileInfoUser0Dir0)
	rep.Create(testFileInfoUser0Dir0File0)
}

func TestCreateFileInfoRepository(t *testing.T) {
	defer testCloseClearGraph()
	testConnectClearGraph()

	_, err := CreateFileInfoRepository()
	if err != nil {
		t.Errorf("Failed to create fileInfo repository: %v", err)
	}

	if t.Failed() {
		testFileInfoSetupFailed = true
	}
}

func TestCreateRootFolder(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	err := rep.CreateUserFolder(testFileInfoUser0.Username)
	if err != nil {
		t.Errorf("Failed to create root folder for user0: %v", err)
	}

	if t.Failed() {
		testFileInfoSetupFailed = true
	}
}

func TestCreateFile(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()
	rep.CreateUserFolder(testFileInfoUser0.Username)

	err := rep.Create(testFileInfoUser0Dir0)
	if err != nil {
		t.Errorf("Failed to create folder: %v", err)
	}
	err = rep.Create(testFileInfoUser0Dir0File0)
	if err != nil {
		t.Errorf("Failed to create file: %v", err)
	}

	if t.Failed() {
		testFileInfoSetupFailed = true
	}
}

func TestGetByPath(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	testFileInfoInsert(rep)

	path := filepath.Join(testFileInfoUser0Dir0.Path, testFileInfoUser0Dir0.Name)
	fileInfo, err := rep.GetByPath(testFileInfoUser0Dir0.OwnerUsername, path)
	if err != nil {
		t.Fatalf("Failed to get file info by path '%s': %v", path, err)
	}
	if !reflect.DeepEqual(fileInfo, testFileInfoUser0Dir0) {
		t.Errorf("Read back file info and file info not deeply equal: %v != %v", fileInfo, testFileInfoUser0Dir0)
	}
}

func TestGetDirectoryContentByPath(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	testFileInfoInsert(rep)

	fileInfos, err := rep.GetDirectoryContentByPath(testFileInfoUser0.Username, testFileInfoUser0Dir0.Path)
	if err != nil {
		t.Fatalf("Failed to get directory content by path '%s': %v", testFileInfoUser0Dir0.Path, err)
	}
	if len(fileInfos) != 1 {
		t.Fatalf("Length of directory contend unequal to one: %d", len(fileInfos))
	}
	if !reflect.DeepEqual(fileInfos[0], testFileInfoUser0Dir0) {
		t.Errorf("Read back directory content and content are not deeply equal: %v != %v", fileInfos[0], testFileInfoUser0Dir0)
	}
}

func TestDeleteFiles(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	testFileInfoInsert(rep)

	err := rep.Delete(testFileInfoUser0.Username, testFileInfoUser0Dir0File0.Path)
	if err != nil {
		t.Fatalf("Failed to delete user0dir0: %v", err)
	}
	count, _ := rep.Count()
	if count != 1 {
		t.Errorf("Count of files after deletion of user0dir0 unequal to one: %d", count)
	}
}

func TestSearchFiles(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	testFileInfoInsert(rep)

	results, err := rep.Search(testFileInfoUser0.Username, testFileInfoUser0Dir0.Path, "0")
	if err != nil {
		t.Fatalf("Failed to search for '%s' in path '%s' for user '%s': %v", "0", testFileInfoUser0Dir0.Path, testFileInfoUser0.Username, err)
	}
	if len(results) != 2 {
		t.Fatalf("Length of result of search unequal to two: %d", len(results))
	}
	if !reflect.DeepEqual(results[0], testFileInfoUser0Dir0) {
		t.Errorf("First search result and user0dir0 are not deeply equal: %v != %v", results[0], testFileInfoUser0Dir0)
	}
	if !reflect.DeepEqual(results[1], testFileInfoUser0Dir0File0) {
		t.Errorf("Second search result and user0dir0file0 are not deeply equal: %v != %v", results[0], testFileInfoUser0Dir0File0)
	}
}

func TestDeleteUserFiles(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	testFileInfoInsert(rep)

	err := rep.DeleteUserFileInfos(testFileInfoUser0.Username)
	if err != nil {
		t.Fatalf("Failed to delete file infos for user0: %v", err)
	}
	count, _ := rep.Count()
	if count != 0 {
		t.Errorf("Count after user file deletion unequal to zero: %d", count)
	}
}

func TestCountFileInfo(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	testFileInfoInsert(rep)

	count, err := rep.Count()
	if err != nil {
		t.Fatalf("Failed to get count of file infos: %v", err)
	}
	if count != 3 {
		t.Errorf("Count of file infos unequal to three: %d", count)
	}
}

func TestGetFSPath(t *testing.T) {
	if testFileInfoSetupFailed {
		t.Skip("Skip due to failed setup")
	}
	defer testCloseClearGraph()
	rep := testFileInfoSetup()

	testFileInfoInsert(rep)

	// TODO: Change to a proper test with different users after implementing sharing
	owner, path, err := rep.GetOwnerPath(testFileInfoUser0.Username, testFileInfoUser0Dir0File0.Path)
	if err != nil {
		t.Fatalf("Failed to get owner path: %v", err)
	}
	if owner != testFileInfoUser0.Username {
		t.Errorf("Read owner of file is unequal to real owner: %v != %v", owner, testFileInfoUser0.Username)
	}
	if path != testFileInfoUser0Dir0File0.Path {
		t.Errorf("Read path of owner of file is unequal to real path of owner: %v != %v", path, testFileInfoUser0Dir0File0.Path)
	}
}
