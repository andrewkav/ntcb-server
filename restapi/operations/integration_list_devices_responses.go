// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"ntcb-server/restmodels"
)

// IntegrationListDevicesOKCode is the HTTP code returned for type IntegrationListDevicesOK
const IntegrationListDevicesOKCode int = 200

/*IntegrationListDevicesOK OK

swagger:response integrationListDevicesOK
*/
type IntegrationListDevicesOK struct {

	/*
	  In: Body
	*/
	Payload []*restmodels.Device `json:"body,omitempty"`
}

// NewIntegrationListDevicesOK creates IntegrationListDevicesOK with default headers values
func NewIntegrationListDevicesOK() *IntegrationListDevicesOK {

	return &IntegrationListDevicesOK{}
}

// WithPayload adds the payload to the integration list devices o k response
func (o *IntegrationListDevicesOK) WithPayload(payload []*restmodels.Device) *IntegrationListDevicesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the integration list devices o k response
func (o *IntegrationListDevicesOK) SetPayload(payload []*restmodels.Device) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *IntegrationListDevicesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*restmodels.Device, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}