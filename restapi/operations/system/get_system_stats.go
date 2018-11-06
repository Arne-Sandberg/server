// Code generated by go-swagger; DO NOT EDIT.

package system

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"

	models "github.com/freecloudio/freecloud/models"
)

// GetSystemStatsHandlerFunc turns a function with the right signature into a get system stats handler
type GetSystemStatsHandlerFunc func(GetSystemStatsParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn GetSystemStatsHandlerFunc) Handle(params GetSystemStatsParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// GetSystemStatsHandler interface for that can handle valid get system stats params
type GetSystemStatsHandler interface {
	Handle(GetSystemStatsParams, *models.Principal) middleware.Responder
}

// NewGetSystemStats creates a new http.Handler for the get system stats operation
func NewGetSystemStats(ctx *middleware.Context, handler GetSystemStatsHandler) *GetSystemStats {
	return &GetSystemStats{Context: ctx, Handler: handler}
}

/*GetSystemStats swagger:route GET /system/stats system getSystemStats

Get system status

*/
type GetSystemStats struct {
	Context *middleware.Context
	Handler GetSystemStatsHandler
}

func (o *GetSystemStats) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetSystemStatsParams()

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
