package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/freecloudio/freecloud/utils"
	"github.com/go-restit/lzjson"

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
		c.Data["response"] = apiModels.Error{Code: http.StatusBadRequest, Message: "No 'files' form field, aborting file upload"}
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
		filePath := filepath.Join(path, files[i].Filename)
		dst, err := s.filesystem.NewFileHandleForUser(user, filePath)
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

		err = s.filesystem.FinishNewFile(user, filePath)
		if err != nil {
			log.Error(0, "Could not finish new file: %v", err)
			c.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	c.WriteHeader(http.StatusCreated)
}

func (s Server) DownloadHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)
	fullPath, filename, err := s.filesystem.GetDownloadPath(user, path)
	if err != nil || filename == "" {
		// TODO: ERROR!
		log.Error(0, "Could not resolve filepath for download: %v", err)
	}
	c.ServeFile(fullPath, filename)
}

func (s Server) ZipHandler(c *macaron.Context) {
	//TODO
	user := c.Data["user"].(*models.User)
	paths := c.Data["request"].(*apiModels.ZipRequest).Paths

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
		ZipPath: utils.ConvertToSlash(fullZipPath, false),
	}
}

func (s Server) StarredFileInfoHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)

	log.Trace("Getting starred fileInfo for %s %s", user.FirstName, user.LastName)

	starredFilesInfo, err := s.filesystem.ListStarredFilesForUser(user)
	if err != nil {
		c.Data["response"] = err
		return
	}

	c.Data["response"] = struct {
		Success          bool               `json:"success"`
		StarredFilesInfo []*models.FileInfo `json:"starred_files_info"`
	}{
		Success:          true,
		StarredFilesInfo: starredFilesInfo,
	}
}

func (s Server) FileInfoHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)

	log.Trace("Getting fileInfo of %s for %s %s", path, user.FirstName, user.LastName)

	fileInfo, content, err := s.filesystem.ListFilesForUser(user, path)
	if err != nil {
		c.Data["response"] = err
		return
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
		err = s.filesystem.FinishNewFile(user, filePath)
		if err != nil {
			c.Data["response"] = err
			return
		}
	}

	fileInfo, err := s.filesystem.GetFileInfo(user, filePath)
	if err != nil {
		c.Data["response"] = err
		return
	}

	c.Data["response"] = struct {
		Success  bool             `json:"success"`
		FileInfo *models.FileInfo `json:"fileInfo"`
	}{
		true,
		fileInfo,
	}

	return
}

func (s Server) UpdateFileHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)
	fileUpdateJSON := c.Data["request"].(lzjson.Node)

	updatedFileInfo, err := s.filesystem.UpdateFile(user, path, fillFileUpdates(fileUpdateJSON))

	if err != nil {
		c.Data["response"] = err
	} else {
		c.Data["response"] = struct {
			Success  bool             `json:"success"`
			FileInfo *models.FileInfo `json:"fileInfo"`
		}{
			true,
			updatedFileInfo,
		}
	}
}

var allowedFileUpdates = []string{
	"path",
	"name",
	"copy",
	"starred",
}

func fillFileUpdates(fileUpdateJSON lzjson.Node) (updates map[string]interface{}) {
	updates = make(map[string]interface{})

	var temp interface{}
	for _, identifier := range allowedFileUpdates {
		value := fileUpdateJSON.Get(identifier)
		if err := value.ParseError(); err != nil {
			continue
		}

		value.Unmarshal(&temp)
		updates[identifier] = temp
	}

	return
}

func (s *Server) FileDeleteHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)
	err := s.filesystem.DeleteFile(user, path)

	if err != nil {
		c.Data["response"] = err
	} else {
		c.Data["response"] = apiModels.SuccessResponse
	}
}

func (s *Server) SearchHandler(c *macaron.Context) {
	user := c.Data["user"].(*models.User)
	path := c.Data["path"].(string)
	results, err := s.filesystem.SearchForFiles(user, path)

	if err != nil {
		c.Data["response"] = err
	} else {
		c.Data["response"] = struct {
			Success bool               `json:"success"`
			Results []*models.FileInfo `json:"results"`
		}{
			true,
			results,
		}
	}
}
