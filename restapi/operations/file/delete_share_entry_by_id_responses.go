// Code generated by go-swagger; DO NOT EDIT.

package file

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/freecloudio/server/models"
)

// DeleteShareEntryByIDOKCode is the HTTP code returned for type DeleteShareEntryByIDOK
const DeleteShareEntryByIDOKCode int = 200

/*DeleteShareEntryByIDOK Success

swagger:response deleteShareEntryByIdOK
*/
type DeleteShareEntryByIDOK struct {
}

// NewDeleteShareEntryByIDOK creates DeleteShareEntryByIDOK with default headers values
func NewDeleteShareEntryByIDOK() *DeleteShareEntryByIDOK {

	return &DeleteShareEntryByIDOK{}
}

// WriteResponse to the client
func (o *DeleteShareEntryByIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

/*DeleteShareEntryByIDDefault Unexpected error

swagger:response deleteShareEntryByIdDefault
*/
type DeleteShareEntryByIDDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteShareEntryByIDDefault creates DeleteShareEntryByIDDefault with default headers values
func NewDeleteShareEntryByIDDefault(code int) *DeleteShareEntryByIDDefault {
	if code <= 0 {
		code = 500
	}

	return &DeleteShareEntryByIDDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the delete share entry by ID default response
func (o *DeleteShareEntryByIDDefault) WithStatusCode(code int) *DeleteShareEntryByIDDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the delete share entry by ID default response
func (o *DeleteShareEntryByIDDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the delete share entry by ID default response
func (o *DeleteShareEntryByIDDefault) WithPayload(payload *models.Error) *DeleteShareEntryByIDDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete share entry by ID default response
func (o *DeleteShareEntryByIDDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteShareEntryByIDDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}