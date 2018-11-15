package restapi

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-openapi/runtime"
	log "gopkg.in/clog.v1"
)

const (
	oneGigabyte = 1024 * 1024 * 1024 * 1024
	fileField   = "upfile"
)

// MultipartformConsumer provides a Consumer for multipart forms.
// As this is only used with file uploads, it will return an io.Reader pointing to
// the data of the <fileField> form element.
func MultipartformConsumer() runtime.Consumer {
	return runtime.ConsumerFunc(func(reader io.Reader, data interface{}) error {
		// Parse the multipart form in the request
		req := http.Request{Body: ioutil.NopCloser(reader)}
		err := req.ParseMultipartForm(5 * oneGigabyte)
		if err != nil {
			return err
		}

		multiform := req.MultipartForm

		// Get the *fileheaders
		files, ok := multiform.File[fileField]
		if !ok {
			log.Error(0, "No '%s' form field, aborting file upload", fileField)
			return fmt.Errorf("missing file field in form")
		}
		if len(files) > 1 {
			log.Error(0, "Uploading more than one file in one request is not allowed")
			return fmt.Errorf("more than one file uploaded in one request")
		}
		file, err := files[0].Open()
		if err != nil {
			log.Error(0, "Error opening file from form: %v", err)
			return err
		}
		data = file
		return nil

	})
}
