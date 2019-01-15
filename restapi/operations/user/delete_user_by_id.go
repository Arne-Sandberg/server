// Code generated by go-swagger; DO NOT EDIT.

package user

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/freecloudio/server/models"
)

// DeleteUserByIDHandlerFunc turns a function with the right signature into a delete user by ID handler
type DeleteUserByIDHandlerFunc func(DeleteUserByIDParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteUserByIDHandlerFunc) Handle(params DeleteUserByIDParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// DeleteUserByIDHandler interface for that can handle valid delete user by ID params
type DeleteUserByIDHandler interface {
	Handle(DeleteUserByIDParams, *models.Principal) middleware.Responder
}

// NewDeleteUserByID creates a new http.Handler for the delete user by ID operation
func NewDeleteUserByID(ctx *middleware.Context, handler DeleteUserByIDHandler) *DeleteUserByID {
	return &DeleteUserByID{Context: ctx, Handler: handler}
}

/*DeleteUserByID swagger:route DELETE /user/{id} user deleteUserById

Delete user by id

*/
type DeleteUserByID struct {
	Context *middleware.Context
	Handler DeleteUserByIDHandler
}

func (o *DeleteUserByID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewDeleteUserByIDParams()

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
