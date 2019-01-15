// Code generated by go-swagger; DO NOT EDIT.

package file

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/freecloudio/server/models"
)

// RescanCurrentUserOKCode is the HTTP code returned for type RescanCurrentUserOK
const RescanCurrentUserOKCode int = 200

/*RescanCurrentUserOK Success

swagger:response rescanCurrentUserOK
*/
type RescanCurrentUserOK struct {
}

// NewRescanCurrentUserOK creates RescanCurrentUserOK with default headers values
func NewRescanCurrentUserOK() *RescanCurrentUserOK {

	return &RescanCurrentUserOK{}
}

// WriteResponse to the client
func (o *RescanCurrentUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

/*RescanCurrentUserDefault Unexpected error

swagger:response rescanCurrentUserDefault
*/
type RescanCurrentUserDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewRescanCurrentUserDefault creates RescanCurrentUserDefault with default headers values
func NewRescanCurrentUserDefault(code int) *RescanCurrentUserDefault {
	if code <= 0 {
		code = 500
	}

	return &RescanCurrentUserDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the rescan current user default response
func (o *RescanCurrentUserDefault) WithStatusCode(code int) *RescanCurrentUserDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the rescan current user default response
func (o *RescanCurrentUserDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the rescan current user default response
func (o *RescanCurrentUserDefault) WithPayload(payload *models.Error) *RescanCurrentUserDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the rescan current user default response
func (o *RescanCurrentUserDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *RescanCurrentUserDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
