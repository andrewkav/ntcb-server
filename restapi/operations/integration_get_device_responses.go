// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"ntcb-server/restmodels"
)

// IntegrationGetDeviceOKCode is the HTTP code returned for type IntegrationGetDeviceOK
const IntegrationGetDeviceOKCode int = 200

/*IntegrationGetDeviceOK OK

swagger:response integrationGetDeviceOK
*/
type IntegrationGetDeviceOK struct {

	/*
	  In: Body
	*/
	Payload *restmodels.Device `json:"body,omitempty"`
}

// NewIntegrationGetDeviceOK creates IntegrationGetDeviceOK with default headers values
func NewIntegrationGetDeviceOK() *IntegrationGetDeviceOK {

	return &IntegrationGetDeviceOK{}
}

// WithPayload adds the payload to the integration get device o k response
func (o *IntegrationGetDeviceOK) WithPayload(payload *restmodels.Device) *IntegrationGetDeviceOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the integration get device o k response
func (o *IntegrationGetDeviceOK) SetPayload(payload *restmodels.Device) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *IntegrationGetDeviceOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
