// Code generated by go-swagger; DO NOT EDIT.

package system

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/freecloudio/server/models"
)

// GetSystemStatsOKCode is the HTTP code returned for type GetSystemStatsOK
const GetSystemStatsOKCode int = 200

/*GetSystemStatsOK System stats

swagger:response getSystemStatsOK
*/
type GetSystemStatsOK struct {

	/*
	  In: Body
	*/
	Payload *models.SystemStats `json:"body,omitempty"`
}

// NewGetSystemStatsOK creates GetSystemStatsOK with default headers values
func NewGetSystemStatsOK() *GetSystemStatsOK {

	return &GetSystemStatsOK{}
}

// WithPayload adds the payload to the get system stats o k response
func (o *GetSystemStatsOK) WithPayload(payload *models.SystemStats) *GetSystemStatsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get system stats o k response
func (o *GetSystemStatsOK) SetPayload(payload *models.SystemStats) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSystemStatsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetSystemStatsDefault Unexpected error

swagger:response getSystemStatsDefault
*/
type GetSystemStatsDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetSystemStatsDefault creates GetSystemStatsDefault with default headers values
func NewGetSystemStatsDefault(code int) *GetSystemStatsDefault {
	if code <= 0 {
		code = 500
	}

	return &GetSystemStatsDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get system stats default response
func (o *GetSystemStatsDefault) WithStatusCode(code int) *GetSystemStatsDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get system stats default response
func (o *GetSystemStatsDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get system stats default response
func (o *GetSystemStatsDefault) WithPayload(payload *models.Error) *GetSystemStatsDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get system stats default response
func (o *GetSystemStatsDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSystemStatsDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
