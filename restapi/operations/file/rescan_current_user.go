// Code generated by go-swagger; DO NOT EDIT.

package file

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/freecloudio/freecloud/models"
)

// RescanCurrentUserHandlerFunc turns a function with the right signature into a rescan current user handler
type RescanCurrentUserHandlerFunc func(RescanCurrentUserParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn RescanCurrentUserHandlerFunc) Handle(params RescanCurrentUserParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// RescanCurrentUserHandler interface for that can handle valid rescan current user params
type RescanCurrentUserHandler interface {
	Handle(RescanCurrentUserParams, *models.Principal) middleware.Responder
}

// NewRescanCurrentUser creates a new http.Handler for the rescan current user operation
func NewRescanCurrentUser(ctx *middleware.Context, handler RescanCurrentUserHandler) *RescanCurrentUser {
	return &RescanCurrentUser{Context: ctx, Handler: handler}
}

/*RescanCurrentUser swagger:route POST /rescan/me file rescanCurrentUser

Rescan own data folder

*/
type RescanCurrentUser struct {
	Context *middleware.Context
	Handler RescanCurrentUserHandler
}

func (o *RescanCurrentUser) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewRescanCurrentUserParams()

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
