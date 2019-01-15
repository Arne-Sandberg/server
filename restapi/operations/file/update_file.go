// Code generated by go-swagger; DO NOT EDIT.

package file

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/freecloudio/server/models"
)

// UpdateFileHandlerFunc turns a function with the right signature into a update file handler
type UpdateFileHandlerFunc func(UpdateFileParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn UpdateFileHandlerFunc) Handle(params UpdateFileParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// UpdateFileHandler interface for that can handle valid update file params
type UpdateFileHandler interface {
	Handle(UpdateFileParams, *models.Principal) middleware.Responder
}

// NewUpdateFile creates a new http.Handler for the update file operation
func NewUpdateFile(ctx *middleware.Context, handler UpdateFileHandler) *UpdateFile {
	return &UpdateFile{Context: ctx, Handler: handler}
}

/*UpdateFile swagger:route PATCH /file file updateFile

Update file/folder

*/
type UpdateFile struct {
	Context *middleware.Context
	Handler UpdateFileHandler
}

func (o *UpdateFile) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewUpdateFileParams()

	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		r = aCtx
	}
	var principal *models.Principal
	if uprinc != nil {
		principal = uprinc.(*models.Principal) // this is really a models.Principal, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
