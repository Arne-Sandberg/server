// Code generated by go-swagger; DO NOT EDIT.

package file

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	models "github.com/freecloudio/server/models"
)

// NewZipFilesParams creates a new ZipFilesParams object
// no default values defined in spec.
func NewZipFilesParams() ZipFilesParams {

	return ZipFilesParams{}
}

// ZipFilesParams contains all the bound params for the zip files operation
// typically these are obtained from a http.Request
//
// swagger:parameters zipFiles
type ZipFilesParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: body
	*/
	Paths *models.PathList
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewZipFilesParams() beforehand.
func (o *ZipFilesParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.PathList
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("paths", "body"))
			} else {
				res = append(res, errors.NewParseError("paths", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Paths = &body
			}
		}
	} else {
		res = append(res, errors.Required("paths", "body"))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
