package repository

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/freecloudio/server/config"
	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/utils"
	"github.com/mholt/archiver"
	log "gopkg.in/clog.v1"
)

const tmpName = ".tmp"

var (
	// ErrForbiddenPathName indicates a path having weird characters that nobody should use, also these characters are forbidden on Windows
	ErrForbiddenPathName = errors.New("paths cannot contain the following characters: <>:\"\\|?*")
	ErrFileNotExist      = errors.New("file does not exist")
)

type FileSystemRepository struct {
	base string
	done chan struct{}
}

func CreateFileSystemRepository(baseDir string, tmpDataExpiry int) (*FileSystemRepository, error) {
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
			return nil, err
		}
	} else if !baseInfo.IsDir() {
		log.Fatal(0, "Base directory does exist but is not a directory")
		return nil, err
	} else if err != nil {
		log.Fatal(0, "Could not check if base directory exists: %v", err)
		return nil, err
	}

	log.Info("Initialized filesystem at base directory %s", base)
	fileSystemRepository := &FileSystemRepository{
		base: base,
		done: make(chan struct{}),
	}

	go fileSystemRepository.cleanupTempFolderRoutine(time.Hour * time.Duration(tmpDataExpiry))

	return fileSystemRepository, nil
}

func (dfs *FileSystemRepository) Close() {
	dfs.done <- struct{}{}
}

func (dfs *FileSystemRepository) cleanupTempFolderRoutine(interval time.Duration) {
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

func (dfs *FileSystemRepository) cleanupTempFolder() {
	log.Trace("Cleaning temp folder")

	infoList, err := ioutil.ReadDir(dfs.base)
	if err != nil {
		log.Warn("Cleaning temp folder failed: %v", err)
	}

	for _, info := range infoList {
		if !info.IsDir() {
			continue
		}

		tmpFolderPath := filepath.Join(dfs.base, info.Name(), tmpName)
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

// NewHandle opens an *os.File handle for writing to.
// Before opening the file, it check the path for sanity.
func (dfs *FileSystemRepository) NewHandle(path string) (*os.File, error) {
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
func (dfs *FileSystemRepository) CreateDirectory(path string) error {
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
func (dfs *FileSystemRepository) CreateDirIfNotExist(path string) (created bool, err error) {
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
func (dfs *FileSystemRepository) GetDirectoryContent(userPath, path string) ([]*models.FileInfo, error) {
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
		fileInfos[i] = dfs.generateInfo(f, path)
	}
	return fileInfos, nil
}

func (dfs *FileSystemRepository) GetInfo(userPath, path, name string) (fileInfo *models.FileInfo, err error) {
	osFileInfo, err := os.Stat(filepath.Join(dfs.base, userPath, path, name))
	if os.IsNotExist(err) {
		err = ErrFileNotExist
		return
	} else if err != nil {
		err = fmt.Errorf("Error resolving file path: %v", err)
		return
	}

	fileInfo = dfs.generateInfo(osFileInfo, path)
	return
}

func (dfs *FileSystemRepository) generateInfo(osFileInfo os.FileInfo, path string) *models.FileInfo {
	return &models.FileInfo{
		Path:        utils.ConvertToSlash(path, true),
		Name:        osFileInfo.Name(),
		IsDir:       osFileInfo.IsDir(),
		Size:        osFileInfo.Size(),
		LastChanged: osFileInfo.ModTime().UTC().Unix(),
		MimeType:    mime.TypeByExtension(filepath.Ext(osFileInfo.Name())),
	}
}

func (dfs *FileSystemRepository) GetDownloadPath(path string) string {
	return filepath.Join(dfs.base, path)
}

// Zip zips all given absolute paths to a zip archive with the given path
func (dfs *FileSystemRepository) Zip(paths []string, outputPath string) (err error) {
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

func (dfs *FileSystemRepository) Move(oldPath, newPath string) (err error) {
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

func (dfs *FileSystemRepository) Delete(path string) (err error) {
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

func (dfs *FileSystemRepository) Copy(oldPath, newPath string) (err error) {
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
