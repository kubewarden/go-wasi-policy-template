package main

import (
	"encoding/json"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewardenProtocol "github.com/kubewarden/policy-sdk-go/protocol"
)

func validate(input []byte) ([]byte, error) {
	validationRequest := kubewardenProtocol.ValidationRequest{}

	err := json.Unmarshal(input, &validationRequest)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(fmt.Sprintf("Error deserializing validation request: %v", err)),
			kubewarden.Code(400))
	}
	settings, err := NewSettingsFromValidationReq(&validationRequest)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(fmt.Sprintf("Error serializing RawMessage: %v", err)),
			kubewarden.Code(400))
	}

	return validateAdmissionReview(settings, validationRequest.Request)
}

func validateAdmissionReview(policySettings Settings, request kubewardenProtocol.KubernetesAdmissionRequest) ([]byte, error) {
	pod := corev1.Pod{}
	err := json.Unmarshal(request.Object, &pod)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(fmt.Sprintf("Error deserializing request object into unstructured: %v", err)),
			kubewarden.Code(400))
	}

	annotations := pod.Metadata.Annotations
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
		return kubewarden.RejectRequest(
			kubewarden.Message(fmt.Sprintf("The following annotations are forbidden: %s", forbiddenAnnotations.String())),
			kubewarden.Code(400))
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
		pod.Metadata.Annotations = annotations
		return kubewarden.MutateRequest(&pod)
	}

	return kubewarden.AcceptRequest()
}
