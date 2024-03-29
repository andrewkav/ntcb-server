// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"
)

// NewIntegrationListDevicesParams creates a new IntegrationListDevicesParams object
// no default values defined in spec.
func NewIntegrationListDevicesParams() IntegrationListDevicesParams {

	return IntegrationListDevicesParams{}
}

// IntegrationListDevicesParams contains all the bound params for the integration list devices operation
// typically these are obtained from a http.Request
//
// swagger:parameters integrationListDevices
type IntegrationListDevicesParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: query
	*/
	Status *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewIntegrationListDevicesParams() beforehand.
func (o *IntegrationListDevicesParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qStatus, qhkStatus, _ := qs.GetOK("status")
	if err := o.bindStatus(qStatus, qhkStatus, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindStatus binds and validates parameter Status from query.
func (o *IntegrationListDevicesParams) bindStatus(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false
	if raw == "" { // empty values pass all other validations
		return nil
	}

	o.Status = &raw

	if err := o.validateStatus(formats); err != nil {
		return err
	}

	return nil
}

// validateStatus carries on validations for parameter Status
func (o *IntegrationListDevicesParams) validateStatus(formats strfmt.Registry) error {

	if err := validate.Enum("status", "query", *o.Status, []interface{}{"online", "offline"}); err != nil {
		return err
	}

	return nil
}
