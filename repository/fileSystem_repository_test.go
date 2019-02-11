package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/freecloudio/server/models"
)

var testFileSystemSetupFailed = false
var testFileSystemDirName = "testData"
var testFileSystemTmpName = ".tmp"

func testFileSystemCleanup(rep *FileSystemRepository) {
	os.RemoveAll(testFileSystemDirName)
	if rep != nil {
		rep.Close()
	}
}

func testFileSystemSetup() *FileSystemRepository {
	testFileSystemCleanup(nil)
	rep, _ := CreateFileSystemRepository(testFileSystemDirName, testFileSystemTmpName, 1, 0)
	return rep
}

func testFileSystemInsertDir(rep *FileSystemRepository) {
	rep.CreateDirectory("1/.tmp")
	rep.CreateDirectory("/2")
}

func testFileSystemInsertFile(rep *FileSystemRepository) {
	file, _ := rep.CreateHandle("1/.tmp/testfile.txt")
	file.Close()
	file, _ = rep.CreateHandle("2/anotherFile.txt")
	file.Close()
}

func testFileSystemInsertComplete(rep *FileSystemRepository) {
	testFileSystemInsertDir(rep)
	testFileSystemInsertFile(rep)
}

func TestCreateFileSystemRepository(t *testing.T) {
	testFileSystemCleanup(nil)

	rep, err := CreateFileSystemRepository(testFileSystemDirName, testFileSystemTmpName, 1, 0)
	if err != nil {
		t.Errorf("Failed to create fileSystemRepository> %v", err)
	}

	if t.Failed() {
		testFileSystemSetupFailed = true
	}

	testFileSystemCleanup(rep)
}

func TestFileSystemClose(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(nil)

	err := rep.Close()
	if err != nil {
		t.Errorf("Failed to close repository: %v", err)
	}

	if t.Failed() {
		testFileSystemSetupFailed = true
	}
}
func TestFileSystemCreateDir(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

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

	if t.Failed() {
		testFileSystemSetupFailed = true
	}

	created, err = rep.CreateDirectory("~/badDir")
	if err != ErrForbiddenPathName {
		t.Errorf("Error for forbidden file name is unequal to ErrForbiddenFileName: %v", err)
	}
	if created {
		t.Error("Directory with forbidden name created")
	}
}

func TestFileSystemCreateFile(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertDir(rep)

	file, err := rep.CreateHandle("1/.tmp/testfile.txt")
	if err != nil {
		t.Errorf("Failed to create new file handle for '1/.tmp/testfile.txt': %v", err)
	}
	file.Close()
	file, err = rep.CreateHandle("2/anotherFile.txt")
	if err != nil {
		t.Errorf("Failed to create file '2/anotherFile.txt': %v", err)
	}
	file.Close()

	if t.Failed() {
		testFileSystemSetupFailed = true
	}

	_, err = rep.CreateHandle("~/badFile.txt")
	if err != ErrForbiddenPathName {
		t.Errorf("Error for forbidden file name is unequal to ErrForbiddenPathName: %v", err)
	}
}

func TestFileSystemGetInfo(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	fileInfo, err := rep.GetInfo("1", ".tmp/testfile.txt")
	if err != nil {
		t.Fatalf("Failed to get fileInfo for '1/.tmp/testfile.txt': %v", err)
	}
	expFileInfo := &models.FileInfo{
		IsDir:    false,
		MimeType: "text/plain; charset=utf-8",
		Name:     "testfile.txt",
		Path:     "/.tmp/",
	}
	fileInfo.LastChanged = 0
	if !reflect.DeepEqual(fileInfo, expFileInfo) {
		t.Errorf("Read fileInfo and expected fileInfo not deeply equal: %v != %v", fileInfo, expFileInfo)
	}
}

func TestFileSystemGetDirectoryInfo(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	dirInfo, err := rep.GetDirectoryInfo("/1", ".tmp")
	if err != nil {
		t.Fatalf("Failed to get directory info for '1/.tmp': %v", err)
	}
	if len(dirInfo) != 1 {
		t.Fatalf("Length of dir info unequal to 1: %d", len(dirInfo))
	}
	expFileInfo := &models.FileInfo{
		IsDir:    false,
		MimeType: "text/plain; charset=utf-8",
		Name:     "testfile.txt",
		Path:     "/.tmp/",
	}
	dirInfo[0].LastChanged = 0
	if !reflect.DeepEqual(dirInfo[0], expFileInfo) {
		t.Errorf("Read fileInfo in dir info and expected fileInfo are not deeply equal: %v != %v", dirInfo[0], expFileInfo)
	}
}

func TestFileSystemGetDownloadPath(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	path := "1/.tmp/testfile.txt"
	downloadPath := rep.GetDownloadPath(path)
	expDownloadPath, _ := filepath.Abs(filepath.Join(testFileSystemDirName, path))
	if downloadPath != expDownloadPath {
		t.Errorf("DownloadPath unequal to expected DownloadPath: %s != %s", downloadPath, expDownloadPath)
	}
}

func TestFileSystemMoveFile(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	err := rep.Move("2/anotherFile.txt", "1/movedFile.txt")
	if err != nil {
		t.Fatalf("Failed to move file '2/anotherFile.txt' to '1/movedFile.txt': %v", err)
	}
	_, err = rep.GetInfo("2", "/anotherFile.txt")
	if err == nil || err != ErrFileNotExist {
		t.Errorf("Getting fileInfo of moved file was successfull or error unequal to 'file not found': %v", err)
	}
	fileInfo, err := rep.GetInfo("1", "/movedFile.txt")
	if err != nil {
		t.Fatalf("Failed to get fileInfo of moved file: %v", err)
	}
	expFileInfo := &models.FileInfo{
		IsDir:    false,
		MimeType: "text/plain; charset=utf-8",
		Name:     "movedFile.txt",
		Path:     "/",
	}
	fileInfo.LastChanged = 0
	if !reflect.DeepEqual(fileInfo, expFileInfo) {
		t.Errorf("FileInfo of moved file and expected fileInfo not deeply equal: %v != %v", fileInfo, expFileInfo)
	}
}

func TestFileSystemCopyFile(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	err := rep.Copy("1/.tmp/testfile.txt", "1/copiedFile.txt")
	if err != nil {
		t.Fatalf("Failed to copy file '1/.tmp/testfile.txt' to '1/copiedFile.txt': %v", err)
	}
	_, err = rep.GetInfo("1", "/.tmp/testfile.txt")
	if err != nil {
		t.Errorf("Failed getting fileInfo for orig file after copying: %v", err)
	}
	fileInfo, err := rep.GetInfo("1", "/copiedFile.txt")
	if err != nil {
		t.Fatalf("Failed to get fileInfo of copied file: %v", err)
	}
	expFileInfo := &models.FileInfo{
		IsDir:    false,
		MimeType: "text/plain; charset=utf-8",
		Name:     "copiedFile.txt",
		Path:     "/",
	}
	fileInfo.LastChanged = 0
	if !reflect.DeepEqual(fileInfo, expFileInfo) {
		t.Errorf("FileInfo of copied file and expected fileInfo not deeply equal: %v != %v", fileInfo, expFileInfo)
	}
}

func TestFileSystemDeleteFile(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	err := rep.Delete("2/anotherFile.txt")
	if err != nil {
		t.Fatalf("Failed to delete file '2/anotherFile.txt': %v", err)
	}
	_, err = rep.GetInfo("2", "/anotherFile.txt")
	if err == nil || err != ErrFileNotExist {
		t.Errorf("Getting fileInfo for deleted file succeeded or error is unequal to 'file does not exist': %v", err)
	}
}

func TestFileSystemCleanTemp(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	err := rep.cleanupTempFolder()
	if err != nil {
		t.Fatalf("Failed to cleanup tmp folder: %v", err)
	}
	_, err = rep.GetInfo("1", ".tmp/testfile.txt")
	if err == nil || err != ErrFileNotExist {
		t.Errorf("Reading tmp file after tmp cleanup successfull or error unequal to 'file does not exist': %v", err)
	}
	_, err = rep.GetInfo("2", "/anotherFile.txt")
	if err != nil {
		t.Errorf("Failed to read normal file after tmp cleanup: %v", err)
	}
}

func TestFileSystemCreateZip(t *testing.T) {
	if testFileSystemSetupFailed {
		t.Skip("Skipped due to failed setup")
	}
	rep := testFileSystemSetup()
	defer testFileSystemCleanup(rep)

	testFileSystemInsertComplete(rep)

	err := rep.Zip([]string{"1/.tmp", "2"}, "2/test.zip")
	if err != nil {
		t.Fatalf("Failed to create zip out of '1/.tmp/' and '2': %v", err)
	}
	fileInfo, err := rep.GetInfo("2", "/test.zip")
	if err != nil {
		t.Fatalf("Failed to get fileInfo of created zip: %v", err)
	}
	expFileInfo := &models.FileInfo{
		IsDir:    false,
		MimeType: "application/zip",
		Name:     "test.zip",
		Path:     "/",
	}
	fileInfo.LastChanged = 0
	fileInfo.Size = 0
	if !reflect.DeepEqual(fileInfo, expFileInfo) {
		t.Errorf("FileInfo of created zip and expected fileInfo not deeply equal: %v", fileInfo)
	}
}
