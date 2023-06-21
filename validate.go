package main

import (
	"encoding/json"
	"fmt"
	"log"

	mapset "github.com/deckarep/golang-set/v2"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func validate(input []byte) []byte {
	var validationRequest ValidationRequest

	err := json.Unmarshal(input, &validationRequest)
	if err != nil {
		return marshalValidationResponseOrFail(
			RejectRequest(
				Message(fmt.Sprintf("Error deserializing validation request: %v", err)),
				Code(400)))
	}
	settingsJSON, err := validationRequest.SettingsRaw.MarshalJSON()
	if err != nil {
		return marshalValidationResponseOrFail(
			RejectRequest(
				Message(fmt.Sprintf("Error serializing RawMessage: %v", err)),
				Code(400)))
	}
	settings := Settings{
		// required to allow marshal
		ForbiddenAnnotations: mapset.NewSet[string](),
	}
	if err := json.Unmarshal(settingsJSON, &settings); err != nil {
		return marshalValidationResponseOrFail(
			RejectRequest(
				Message(fmt.Sprintf("Error deserializing Settings: %v", err)),
				Code(400)))
	}

	return marshalValidationResponseOrFail(
		validateAdmissionReview(settings, validationRequest.Request))
}

func marshalValidationResponseOrFail(response ValidationResponse) []byte {
	responseBytes, err := json.Marshal(&response)
	if err != nil {
		log.Fatalf("cannot marshal validation response: %v", err)
	}
	return responseBytes
}

func validateAdmissionReview(policySettings Settings, request admissionv1.AdmissionRequest) ValidationResponse {
	obj := unstructured.Unstructured{}
	err := json.Unmarshal(request.Object.Raw, &obj)
	if err != nil {
		return RejectRequest(
			Message(fmt.Sprintf("Error deserializing request object into unstructured: %v", err)),
			Code(400))
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// check if one of the annotations is forbidden
	annotationsSet := mapset.NewSet[string]()
	for key := range annotations {
		annotationsSet.Add(key)
	}
	forbiddenAnnotations := annotationsSet.Intersect(policySettings.ForbiddenAnnotations)
	if forbiddenAnnotations.Cardinality() > 0 {
		return RejectRequest(
			Message(fmt.Sprintf("The following annotations are forbidden: %s", forbiddenAnnotations.String())),
			Code(400))
	}

	// eventually mutate the current annotations
	annotationsChanged := false
	for key, value := range policySettings.RequiredAnnotations {
		currentValue, hasKey := annotations[key]
		if !hasKey || currentValue != value {
			annotations[key] = value
			annotationsChanged = true
		}
	}

	if annotationsChanged {
		obj.SetAnnotations(annotations)
		return MutateRequest(&obj)
	}

	return AcceptRequest()
}
