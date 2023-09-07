package main

//nolint:godox
// TODO: figure out if it's worth to move this to a dedicated library
// We cannot add that to the Kubewarden Go SDK because it would not work
// with TinyGo

import (
	"encoding/json"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ValidationRequest describes the input received by the policy
// when invoked via the `validate` subcommand
type ValidationRequest struct {
	Request     admissionv1.AdmissionRequest `json:"request"`
	SettingsRaw json.RawMessage              `json:"settings"`
}

// Message is the optional string used to build validation responses
type Message string

// Code is the optional error code associated with validation responses
type Code uint16

const (
	// NoMessage can be used when building a response that doesn't have any
	// message to be shown to the user
	NoMessage Message = ""

	// NoCode can be used when building a response that doesn't have any
	// error code to be shown to the user
	NoCode Code = 0
)

// ValidationResponse defines the response given when validating a request
//
//nolint:tagliatelle
type ValidationResponse struct {
	Accepted bool `json:"accepted"`
	// Optional - ignored if accepted
	Message *string `json:"message,omitempty"`
	// Optional - ignored if accepted
	Code *uint16 `json:"code,omitempty"`
	// Optional - used only by mutating policies
	MutatedObject *unstructured.Unstructured `json:"mutated_object,omitempty"`
}

// SettingsValidationResponse is the response sent by a policy when validating
// its settings
type SettingsValidationResponse struct {
	Valid bool `json:"valid"`
	// Optional - ignored if valid
	Message *string `json:"message,omitempty"`
}

// AcceptRequest can be used inside of the `validate` function to accept the
// incoming request
func AcceptRequest() ValidationResponse {
	return ValidationResponse{
		Accepted: true,
	}
}

// MutateRequest accepts the request. The given `mutatedObject` is how
// the evaluated object must look once accepted
func MutateRequest(mutatedObject *unstructured.Unstructured) ValidationResponse {
	return ValidationResponse{
		Accepted:      true,
		MutatedObject: mutatedObject,
	}
}

// RejectRequest can be used inside of the `validate` function to reject the
// incoming request
// * `message`: optional message to show to the user
// * `code`: optional error code to show to the user
func RejectRequest(message Message, code Code) ValidationResponse {
	response := ValidationResponse{
		Accepted: false,
	}
	if message != NoMessage {
		msg := string(message)
		response.Message = &msg
	}
	if code != NoCode {
		c := uint16(code)
		response.Code = &c
	}

	return response
}

// AcceptSettings be used inside of the `validateSettings` function to accept the
// incoming settings
func AcceptSettings() SettingsValidationResponse {
	return SettingsValidationResponse{
		Valid: true,
	}
}

// RejectSettings can be used inside of the `validate_settings` function to
// mark the user provided settings as invalid
// * `message`: optional message to show to the user
func RejectSettings(message Message) SettingsValidationResponse {
	response := SettingsValidationResponse{
		Valid: false,
	}

	if message != NoMessage {
		msg := string(message)
		response.Message = &msg
	}

	return response
}
