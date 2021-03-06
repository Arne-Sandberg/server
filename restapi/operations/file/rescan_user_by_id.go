// Code generated by go-swagger; DO NOT EDIT.

package file

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/freecloudio/server/models"
)

// RescanUserByIDHandlerFunc turns a function with the right signature into a rescan user by ID handler
type RescanUserByIDHandlerFunc func(RescanUserByIDParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn RescanUserByIDHandlerFunc) Handle(params RescanUserByIDParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// RescanUserByIDHandler interface for that can handle valid rescan user by ID params
type RescanUserByIDHandler interface {
	Handle(RescanUserByIDParams, *models.Principal) middleware.Responder
}

// NewRescanUserByID creates a new http.Handler for the rescan user by ID operation
func NewRescanUserByID(ctx *middleware.Context, handler RescanUserByIDHandler) *RescanUserByID {
	return &RescanUserByID{Context: ctx, Handler: handler}
}

/*RescanUserByID swagger:route POST /file/rescan/{id} file rescanUserById

Rescan data folder by user id

*/
type RescanUserByID struct {
	Context *middleware.Context
	Handler RescanUserByIDHandler
}

func (o *RescanUserByID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewRescanUserByIDParams()

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
