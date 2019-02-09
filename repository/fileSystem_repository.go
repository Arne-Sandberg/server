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

	"github.com/freecloudio/server/models"
	"github.com/freecloudio/server/utils"
	"github.com/mholt/archiver"
	log "gopkg.in/clog.v1"
)

var (
	// ErrForbiddenPathName indicates a path having weird characters that nobody should use, also these characters are forbidden on Windows
	ErrForbiddenPathName = errors.New("paths cannot contain the following characters: <>:\"\\|?*")
	// ErrFileNotExist is the error that a file does not exist
	ErrFileNotExist = errors.New("file does not exist")
)

// FileSystemRepository represents the local filesystem for storing files
type FileSystemRepository struct {
	base    string
	tmpName string
	done    chan struct{}
}

// CreateFileSystemRepository creates a new fileSystemRepository at a given relative or abolute path with a interval for temp cleanup in hours and a tmp data expiry in hours
func CreateFileSystemRepository(baseDir, tmpName string, tmpClearInterval, tmpDataExpiry int) (*FileSystemRepository, error) {
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

	go fileSystemRepository.cleanupTempFolderRoutine(tmpClearInterval, tmpDataExpiry)

	return fileSystemRepository, nil
}

// Close closes the repository and with that ends the go routine for tmp cleanup
func (rep *FileSystemRepository) Close() error {
	rep.done <- struct{}{}
	return nil
}

// cleanupTempFolderRoutine is the actual routine that periodically calls cleanupTempFolder
func (rep *FileSystemRepository) cleanupTempFolderRoutine(interval, expiry int) {
	log.Trace("Starting temp folder cleaner, running now and every %v hours", interval)
	rep.cleanupTempFolder(expiry)

	ticker := time.NewTicker(time.Hour * time.Duration(interval))
	for {
		select {
		case <-rep.done:
			return
		case <-ticker.C:
			rep.cleanupTempFolder(expiry)
		}
	}
}

// cleanupTempFolder deletes the content of all tmp folders
func (rep *FileSystemRepository) cleanupTempFolder(expiry int) {
	log.Trace("Cleaning temp folder")

	now := time.Now()

	infoList, err := ioutil.ReadDir(rep.base)
	if err != nil {
		log.Warn("Cleaning temp folder failed: %v", err)
	}

	for _, info := range infoList {
		if !info.IsDir() {
			continue
		}

		tmpFolderPath := filepath.Join(rep.base, info.Name(), rep.tmpName)
		tmpInfoList, err := ioutil.ReadDir(tmpFolderPath)
		if err != nil {
			log.Warn("Error reading temp folder in %v during temp cleanup: %v", tmpFolderPath, err)
		}

		for _, tmpInfo := range tmpInfoList {
			expires := tmpInfo.ModTime().Add(time.Hour * time.Duration(expiry))
			if now.After(expires) {
				err = os.RemoveAll(filepath.Join(tmpFolderPath, tmpInfo.Name()))
				if err != nil {
					log.Warn("Error deleting file %v in temp folder %v during temp cleanup: %v", tmpInfo.Name(), tmpFolderPath, err)
					continue
				}
			}
		}
	}
}

// CreateHandle opens an *os.File handle for writing to.
// Before opening the file, it check the path for sanity.
func (rep *FileSystemRepository) CreateHandle(path string) (*os.File, error) {
	if !utils.ValidatePath(path) {
		return nil, ErrForbiddenPathName
	}
	f, err := os.Create(filepath.Join(rep.base, path))
	if err != nil {
		log.Error(0, "Could not create file %s: %v", path, err)
		return nil, err
	}
	return f, nil
}

// CreateDirectory checks whether directory exists and creates it otherwise
func (rep *FileSystemRepository) CreateDirectory(path string) (created bool, err error) {
	if !utils.ValidatePath(path) {
		return false, ErrForbiddenPathName
	}

	_, fileErr := os.Stat(filepath.Join(rep.base, path))
	if os.IsNotExist(fileErr) {
		log.Info("Directory does not exist, creating it now")
		err := os.MkdirAll(filepath.Join(rep.base, path), 0755)
		if err != nil {
			log.Error(0, "Could not create directory %s: %v", path, err)
		}
		return true, err
	} else if fileErr != nil {
		err = fileErr
		log.Warn("Could not check if directory exists, assuming it does: %v", err)
		return false, err
	}
	return false, nil
}

// GetDirectoryInfo returns a list of all files and folders in the given "path" (relative to the user's directory).
// Before doing so, it checks the path for sanity.
func (rep *FileSystemRepository) GetDirectoryInfo(userPath, path string) ([]*models.FileInfo, error) {
	if !utils.ValidatePath(path) {
		return nil, ErrForbiddenPathName
	}

	info, err := ioutil.ReadDir(filepath.Join(rep.base, userPath, path))
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
		fileInfos[i] = rep.generateInfo(f, path)
	}
	return fileInfos, nil
}

// GetInfo generates and returns the fileInfo of a file in the fileSystem
func (rep *FileSystemRepository) GetInfo(userPath, path, name string) (fileInfo *models.FileInfo, err error) {
	osFileInfo, err := os.Stat(filepath.Join(rep.base, userPath, path, name))
	if os.IsNotExist(err) {
		err = ErrFileNotExist
		return
	} else if err != nil {
		err = fmt.Errorf("Error resolving file path: %v", err)
		return
	}

	fileInfo = rep.generateInfo(osFileInfo, path)
	return
}

func (rep *FileSystemRepository) generateInfo(osFileInfo os.FileInfo, path string) *models.FileInfo {
	return &models.FileInfo{
		Path:        utils.ConvertToSlash(path, true),
		Name:        osFileInfo.Name(),
		IsDir:       osFileInfo.IsDir(),
		Size:        osFileInfo.Size(),
		LastChanged: osFileInfo.ModTime().UTC().Unix(),
		MimeType:    mime.TypeByExtension(filepath.Ext(osFileInfo.Name())),
	}
}

// GetDownloadPath returns the absolute path for the server to serve from
func (rep *FileSystemRepository) GetDownloadPath(path string) string {
	return filepath.Join(rep.base, path)
}

// Zip zips all given absolute paths to a zip archive with the given path
func (rep *FileSystemRepository) Zip(paths []string, outputPath string) (err error) {
	for it := 0; it < len(paths); it++ {
		paths[it] = filepath.Join(rep.base, paths[it])
	}

	fullZipPath := filepath.Join(rep.base, outputPath)
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

func (rep *FileSystemRepository) Move(oldPath, newPath string) (err error) {
	if !utils.ValidatePath(oldPath) {
		err = ErrForbiddenPathName
		return
	}
	if !utils.ValidatePath(newPath) {
		err = ErrForbiddenPathName
		return
	}

	err = os.Rename(filepath.Join(rep.base, oldPath), filepath.Join(rep.base, newPath))
	if err != nil {
		log.Error(0, "Moving %v to %v failed", oldPath, newPath)
		return
	}

	return
}

func (rep *FileSystemRepository) Delete(path string) (err error) {
	if !utils.ValidatePath(path) {
		err = ErrForbiddenPathName
		return
	}

	err = os.RemoveAll(filepath.Join(rep.base, path))
	if err != nil {
		log.Error(0, "Deleting %v failed", path)
		return
	}
	return
}

func (rep *FileSystemRepository) Copy(oldPath, newPath string) (err error) {
	oldFullPath := filepath.Join(rep.base, oldPath)
	newFullPath := filepath.Join(rep.base, newPath)

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
