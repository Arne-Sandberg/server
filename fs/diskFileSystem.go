package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/utils"

	"github.com/freecloudio/freecloud/models"
	"github.com/mholt/archiver"
	log "gopkg.in/clog.v1"
)

// DiskFilesystem implements the Filesystem interface, writing actual files to the disk
type DiskFilesystem struct {
	base    string
	tmpName string
	done    chan struct{}
}

// NewDiskFilesystem sets up a disk filesystem and returns it
func NewDiskFilesystem(baseDir, tmpName string, tmpDataExpiry int) (dfs *DiskFilesystem, err error) {
	base, err := filepath.Abs(baseDir)
	if err != nil {
		log.Error(0, "Could not initialize filesystem: %v", err)
		return nil, err
	}

	// Check if the base directory does not exist. If it doesn't, create it.
	baseInfo, err := os.Stat(base)
	if os.IsNotExist(err) {
		log.Info("Base directory does not exist, creating it now")
		err = os.Mkdir(base, 0755)
		if err != nil {
			log.Error(0, "Could not create base directory: %v", err)
			return
		}
	} else if !baseInfo.IsDir() {
		log.Fatal(0, "Base directory does exist but is not a directory")
		return
	} else if err != nil {
		log.Fatal(0, "Could not check if base directory exists: %v", err)
		return
	}

	log.Info("Initialized filesystem at base directory %s", base)
	dfs = &DiskFilesystem{base: base, tmpName: tmpName, done: make(chan struct{})}

	go dfs.cleanupTempFolderRoutine(time.Hour * time.Duration(tmpDataExpiry))

	return
}

func (dfs *DiskFilesystem) Close() {
	dfs.done <- struct{}{}
}

func (dfs *DiskFilesystem) cleanupTempFolderRoutine(interval time.Duration) {
	log.Trace("Starting temp folder cleaner, running now and every %v", interval)
	dfs.cleanupTempFolder()

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-dfs.done:
			return
		case <-ticker.C:
			dfs.cleanupTempFolder()
		}
	}
}

func (dfs *DiskFilesystem) cleanupTempFolder() {
	log.Trace("Cleaning temp folder")

	infoList, err := ioutil.ReadDir(dfs.base)
	if err != nil {
		log.Warn("Cleaning temp folder failed: %v", err)
	}

	for _, info := range infoList {
		if !info.IsDir() {
			continue
		}

		tmpFolderPath := filepath.Join(dfs.base, info.Name(), dfs.tmpName)
		tmpInfoList, err := ioutil.ReadDir(tmpFolderPath)
		if err != nil {
			log.Warn("Error reading temp folder in %v during temp cleanup: %v", tmpFolderPath, err)
		}

		for _, tmpInfo := range tmpInfoList {
			if time.Now().After(tmpInfo.ModTime().Add(time.Hour * time.Duration(config.GetInt("fs.tmp_data_expiry")))) {
				err = os.RemoveAll(filepath.Join(tmpFolderPath, tmpInfo.Name()))
				if err != nil {
					log.Warn("Error deleting file %v in temp folder %v during temp cleanup: %v", tmpInfo.Name(), tmpFolderPath, err)
					continue
				}
			}
		}
	}
}

// NewFileHandle opens an *os.File handle for writing to.
// Before opening the file, it check the path for sanity.
func (dfs *DiskFilesystem) NewFileHandle(path string) (*os.File, error) {
	if !utils.ValidatePath(path) {
		return nil, ErrForbiddenPathName
	}
	f, err := os.Create(filepath.Join(dfs.base, path))
	if err != nil {
		log.Error(0, "Could not create file %s: %v", path, err)
		return nil, err
	}
	return f, nil
}

// CreateDirectory creates a new directory at "path".
// Before doing so, it check the path for sanity.
func (dfs *DiskFilesystem) CreateDirectory(path string) error {
	log.Trace("Path for new directory is '%s'", path)
	if !utils.ValidatePath(path) {
		return ErrForbiddenPathName
	}
	err := os.MkdirAll(filepath.Join(dfs.base, path), 0755)
	if err != nil {
		log.Error(0, "Could not create directory %s: %v", path, err)
	}
	return err
}

// CreateDirIfNotExist checks whether directory exists and creates it otherwise
func (dfs *DiskFilesystem) CreateDirIfNotExist(path string) (created bool, err error) {
	_, fileErr := os.Stat(filepath.Join(dfs.base, path))
	if os.IsNotExist(fileErr) {
		log.Info("Directory does not exist, creating it now")
		err = dfs.CreateDirectory(path)
		return true, err
	} else if fileErr != nil {
		err = fileErr
		log.Warn("Could not check if directory exists, assuming it does: %v", err)
		return false, err
	}
	return false, nil
}

// GetDirectoryContent returns a list of all files and folders in the given "path" (relative to the user's directory).
// Before doing so, it checks the path for sanity.
func (dfs *DiskFilesystem) GetDirectoryContent(userPath, path string) ([]*models.FileInfo, error) {
	if !utils.ValidatePath(path) {
		return nil, ErrForbiddenPathName
	}

	info, err := ioutil.ReadDir(filepath.Join(dfs.base, userPath, path))
	if err != nil {
		log.Error(0, "Could not list files in %s: %v", path, err)
		return nil, err
	}

	if path == "" {
		path = "/"
	}
	path = utils.ConvertToSlash(path, true)

	fileInfos := make([]*models.FileInfo, len(info), len(info))
	for i, f := range info {
		fileInfos[i] = dfs.generateFileInfo(f, path)
	}
	return fileInfos, nil
}

func (dfs *DiskFilesystem) GetFileInfo(userPath, path, name string) (fileInfo *models.FileInfo, err error) {
	osFileInfo, err := os.Stat(filepath.Join(dfs.base, userPath, path, name))
	if os.IsNotExist(err) {
		err = ErrFileNotExist
		return
	} else if err != nil {
		err = fmt.Errorf("Error resolving file path: %v", err)
		return
	}

	fileInfo = dfs.generateFileInfo(osFileInfo, path)
	return
}

func (dfs *DiskFilesystem) generateFileInfo(osFileInfo os.FileInfo, path string) *models.FileInfo {
	return &models.FileInfo{
		Path:        utils.ConvertToSlash(path, true),
		Name:        osFileInfo.Name(),
		IsDir:       osFileInfo.IsDir(),
		Size:        osFileInfo.Size(),
		LastChanged: utils.GetTimestampFromTime(osFileInfo.ModTime()),
		MimeType:    mime.TypeByExtension(filepath.Ext(osFileInfo.Name())),
	}
}

func (dfs *DiskFilesystem) GetDownloadPath(path string) string {
	return filepath.Join(dfs.base, path)
}

// ZipFiles zips all given absolute paths to a zip archive with the given path
func (dfs *DiskFilesystem) ZipFiles(paths []string, outputPath string) (err error) {
	for it := 0; it < len(paths); it++ {
		paths[it] = filepath.Join(dfs.base, paths[it])
	}

	fullZipPath := filepath.Join(dfs.base, outputPath)
	if err != nil {
		return
	}

	err = archiver.Zip.Make(fullZipPath, paths)
	if err != nil {
		log.Error(0, "Error zipping files into %v: %v", outputPath, err)
		return
	}
	return
}

func (dfs *DiskFilesystem) MoveFile(oldPath, newPath string) (err error) {
	if !utils.ValidatePath(oldPath) {
		err = ErrForbiddenPathName
		return
	}
	if !utils.ValidatePath(newPath) {
		err = ErrForbiddenPathName
		return
	}

	err = os.Rename(filepath.Join(dfs.base, oldPath), filepath.Join(dfs.base, newPath))
	if err != nil {
		log.Error(0, "Moving %v to %v failed", oldPath, newPath)
		return
	}

	return
}

func (dfs *DiskFilesystem) DeleteFile(path string) (err error) {
	if !utils.ValidatePath(path) {
		err = ErrForbiddenPathName
		return
	}

	err = os.RemoveAll(filepath.Join(dfs.base, path))
	if err != nil {
		log.Error(0, "Deleting %v failed", path)
		return
	}
	return
}

func (dfs *DiskFilesystem) CopyFile(oldPath, newPath string) (err error) {
	oldFullPath := filepath.Join(dfs.base, oldPath)
	newFullPath := filepath.Join(dfs.base, newPath)

	in, err := os.Open(oldFullPath)
	if err != nil {
		err = fmt.Errorf("Error opening file %v to copy", oldPath)
		return
	}
	defer in.Close()

	out, err := os.Create(newFullPath)
	if err != nil {
		err = fmt.Errorf("Error creating file %v to copy to", newPath)
		log.Error(0, "%v", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		err = fmt.Errorf("Error copying %v to %v", oldPath, newPath)
		log.Error(0, "%v", err)
		return
	}

	err = out.Sync()
	if err != nil {
		err = fmt.Errorf("Error syncing copied file%v", newPath)
		log.Error(0, "%v", err)
		return
	}
	return
}
