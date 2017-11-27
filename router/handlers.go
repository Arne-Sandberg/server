package router

import (
	"io"
	"net/http"

	"github.com/riesinger/freecloud/config"
	macaron "gopkg.in/macaron.v1"
)

const oneGigabyte = 1024 * 1024 * 1024 * 1024

func (server) PingHandler(c *macaron.Context) {
	c.Write([]byte("Pong"))
}

func (server) FileUpload(c *macaron.Context) {
	if c.Req.Method == http.MethodGet {
		c.HTML(http.StatusOK, "files/upload", nil)
	} else if c.Req.Method == http.MethodPost {
		// Parse the multipart form in the request
		err := c.Req.ParseMultipartForm(config.GetInt64("http.upload_limit") * oneGigabyte)
		if err != nil {
			c.Error(http.StatusInternalServerError, "File upload failed:", err.Error())
			return
		}

		multiform := c.Req.MultipartForm

		// Get the *fileheaders
		files := multiform.File["files"]
		for i := range files {
			// For each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				c.Error(http.StatusInternalServerError, "Could not open file:", err.Error())
				return
			}

			// Create the destination file making sure the path is writeable.
			dst, err := filesystem.NewFileHandle(files[i].Filename)
			defer dst.Close()
			if err != nil {
				c.Error(http.StatusInternalServerError, "Could not open file for writing:", err.Error())
				return
			}

			// Copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				c.Error(http.StatusInternalServerError, "Could not copy the file:", err.Error())
				return
			}
		}

		c.HTML(http.StatusCreated, "files/upload", "Upload successful!")
	}
}
