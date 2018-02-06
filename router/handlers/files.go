package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/freecloudio/freecloud/utils"

	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/models"
	apiModels "github.com/freecloudio/freecloud/models/api"
	log "gopkg.in/clog.v1"
	macaron "gopkg.in/macaron.v1"
)

const oneGigabyte = 1024 * 1024 * 1024 * 1024

func (s Server) UploadHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)

	// Parse the multipart form in the request
	err := c.Req.ParseMultipartForm(config.GetInt64("http.upload_limit") * oneGigabyte)
	if err != nil {
		log.Error(0, "File upload failed: %v", err)
		c.Data["response"] = fmt.Errorf("File upload failed: %v", err)
		return
	}

	multiform := c.Req.MultipartForm

	// Get the *fileheaders
	files, ok := multiform.File["files"]
	if !ok {
		log.Error(0, "No 'files' form field, aborting file upload")
		c.Data["response"] = models.APIError{Code: http.StatusBadRequest, Message: "No 'files' form field, aborting file upload"}
		return
	}
	for i := range files {
		// For each fileheader, get a handle to the actual file
		file, err := files[i].Open()
		if err != nil {
			log.Error(0, "Could not open file: %v", err)
			c.Data["response"] = fmt.Errorf("Could not open file: %v", err)
			return
		}
		defer file.Close()

		// Create the destination file making sure the path is writeable.
		dst, err := s.filesystem.NewFileHandleForUser(user, filepath.Join(path, files[i].Filename))
		if err != nil {
			log.Error(0, "Could not open file for writing: %v", err)
			c.Data["response"] = fmt.Errorf("Could not open file for writing: %v", err)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			log.Error(0, "Could not copy the file: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	c.WriteHeader(http.StatusCreated)
}

func (s Server) DownloadHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)
	fullPath, filename, err := s.filesystem.ResolveFilePath(user, path)
	if err != nil || filename == "" {
		// TODO: ERROR!
		log.Error(0, "Could not resolve filepath for download: %v", err)
	}
	c.ServeFile(fullPath, filename)
}

func (s Server) ZipHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	paths := c.Data["request"].(*apiModels.ZipRequest).Paths

	var err error
	for _, path := range paths {
		path, _, err = s.filesystem.ResolveFilePath(user, path)
		if err != nil {
			c.Data["response"] = err
			return
		}
	}

	outputFileName := "_" + time.Now().UTC().Format("06-01-02_15-04-05") + ".zip"
	if len(paths) == 1 {
		outputFileName = filepath.Base(paths[0]) + outputFileName
	} else {
		outputFileName = "fc" + outputFileName
	}

	fullZipPath, err := s.filesystem.ZipFiles(user, paths, outputFileName)
	if err != nil {
		c.Data["response"] = err
		return
	}

	c.Data["response"] = struct {
		Success bool   `json:"success"`
		ZipPath string `json:"zipPath"`
	}{
		Success: true,
		ZipPath: utils.ConvertToSlash(fullZipPath),
	}
}

func (s Server) FileInfoHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)

	log.Trace("Getting fileInfo of %s for %s %s", path, user.FirstName, user.LastName)

	fileInfo, err := s.filesystem.GetFileInfo(user, path)
	if err != nil {
		c.Data["response"] = err
		return
	}

	var content []*models.FileInfo
	if fileInfo.IsDir {
		content, err = s.filesystem.ListFilesForUser(user, path)
		if err != nil {
			c.Data["response"] = err
			return
		}
	}

	c.Data["response"] = struct {
		Success  bool               `json:"success"`
		FileInfo *models.FileInfo   `json:"fileInfo"`
		Content  []*models.FileInfo `json:"content"`
	}{
		Success:  true,
		FileInfo: fileInfo,
		Content:  content,
	}
	return
}

func (s Server) CreateFileHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)
	fileInfo := c.Data["request"].(*models.FileInfo)
	filePath := filepath.Join(path, fileInfo.Name)

	log.Trace("Creating file '%s' for %s %s", filePath, user.FirstName, user.LastName)

	if fileInfo.IsDir {
		err := s.filesystem.CreateDirectoryForUser(user, filePath)
		// TODO: match agains path errors and return a http.StatusBadRequest on those
		if err != nil {
			c.Data["response"] = err
			return
		}
	} else {
		file, err := s.filesystem.NewFileHandleForUser(user, filePath)
		defer file.Close()
		if err != nil {
			c.Data["response"] = err
			return
		}
	}
	c.Data["response"] = models.SuccessResponse
	return
}

func (s Server) RenameDirectoryHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path, err := url.PathUnescape(c.Params("*"))
	if err != nil {
		c.Data["response"] = fmt.Errorf("Invalid directory format")
	}
	newPath := c.Data["request"].(*models.FileInfo).Path

	log.Trace("Renaming directory '%s' to '%s' for %s %s", path, newPath, user.FirstName, user.LastName)

	// Check for invalid directory (../1/* --> move to other peoples files)
	// Do rename/move
	// Forbid changes too root directory
}
