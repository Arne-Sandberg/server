package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/riesinger/freecloud/config"
	"github.com/riesinger/freecloud/models"
	log "gopkg.in/clog.v1"
	macaron "gopkg.in/macaron.v1"
)

const oneGigabyte = 1024 * 1024 * 1024 * 1024

func (s Server) UploadHandler(c *macaron.Context) {
	// TODO: check if the user actually exists
	user := c.Data["user"].(*models.User)
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
		dst, err := s.filesystem.NewFileHandleForUser(user, files[i].Filename)
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

func (s Server) GetDirectoryHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path, err := url.PathUnescape(c.Params("*"))
	if err != nil {
		c.Data["response"] = fmt.Errorf("Invalid directory format")
	}
	log.Trace("Getting directory contents of %s for %s %s", path, user.FirstName, user.LastName)
	fileInfos, err := s.filesystem.ListFilesForUser(user, path)
	if err != nil {
		c.Data["response"] = err
		return
	}
	c.Data["response"] = struct {
		Success bool               `json:"success"`
		Files   []*models.FileInfo `json:"files"`
	}{
		Success: true,
		Files:   fileInfos,
	}
	return
}
