package main

import (
	"encoding/json"
	"log"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	apimachinery "github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1"
	kubewardenProtocol "github.com/kubewarden/policy-sdk-go/protocol"
)

func TestValidateAdmissionReview(t *testing.T) {
	cases := []struct {
		name                 string
		currentAnnotations   map[string]string
		requiredAnnotations  map[string]string
		forbiddenAnnotations mapset.Set[string]
		isAccepted           bool
		isMutated            bool
	}{
		{
			"object has already the required annotations",
			map[string]string{
				"cc-center": "marketing",
			},
			map[string]string{
				"cc-center": "marketing",
			},
			mapset.NewSet[string](),
			true,
			false,
		},
		{
			"object has a forbidden annotation",
			map[string]string{
				"team": "marketing",
			},
			map[string]string{
				"cc-center": "marketing",
			},
			mapset.NewSet[string]("team"),
			false,
			false,
		},
		{
			"mutate object - add key",
			map[string]string{},
			map[string]string{
				"cc-center": "marketing",
			},
			mapset.NewSet[string]("team"),
			true,
			true,
		},
		{
			"mutate object - update key",
			map[string]string{
				"cc-center": "foo",
			},
			map[string]string{
				"cc-center": "marketing",
			},
			mapset.NewSet[string]("team"),
			true,
			true,
		},
	}

	for _, testCase := range cases {
		settings := Settings{
			RequiredAnnotations:  testCase.requiredAnnotations,
			ForbiddenAnnotations: testCase.forbiddenAnnotations,
		}

		obj := buildPodJSON(testCase.currentAnnotations)

		t.Run(testCase.name, func(t *testing.T) {
			admissionRequest := kubewardenProtocol.KubernetesAdmissionRequest{
				Object: obj,
			}

			responseJSON, err := validateAdmissionReview(settings, admissionRequest)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			response := kubewardenProtocol.ValidationResponse{}
			if err = json.Unmarshal(responseJSON, &response); err != nil {
				t.Errorf("cannot unmarshal validation response: %v", err)
			}

			if response.Accepted != testCase.isAccepted {
				t.Errorf(
					"didn't get the expected validation outcome, %v was expected, got %v instead",
					testCase.isAccepted, response.Accepted)
				if response.Message != nil {
					t.Errorf(
						"policy message: %s",
						*response.Message)
				}
			}
			if response.MutatedObject == nil && testCase.isMutated {
				t.Errorf("object has not been mutated")
			}
			if response.MutatedObject != nil && !testCase.isMutated {
				t.Errorf("object should not have been mutated")
			}
		})
	}
}

func buildPodJSON(annotations map[string]string) json.RawMessage {
	metadata := apimachinery.ObjectMeta{
		Name:        "test",
		Namespace:   "default",
		Annotations: annotations,
	}
	pod := corev1.Pod{
		Metadata: &metadata,
	}
	podJSON, err := json.Marshal(&pod)
	if err != nil {
		log.Fatalf("cannot marshall namespace: %v", err)
	}

	return podJSON
}
