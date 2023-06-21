package main

import (
	"encoding/json"
	"log"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestValidateAdmissionReview(t *testing.T) {
	cases := []struct {
		desc                 string
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

	for _, tc := range cases {
		settings := Settings{
			RequiredAnnotations:  tc.requiredAnnotations,
			ForbiddenAnnotations: tc.forbiddenAnnotations,
		}

		jsonObjects := map[string][]byte{
			"namespace": buildNamespaceJSON(tc.currentAnnotations),
			"service":   buildServiceJSON(tc.currentAnnotations),
		}

		for objType, objJSON := range jsonObjects {
			admissionRequest := admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: objJSON,
				},
			}

			response := validateAdmissionReview(settings, admissionRequest)
			if response.Accepted != tc.isAccepted {
				t.Errorf(
					"[%s/%s] didn't get the expected validation outcome, %v was expected, got %v instead",
					tc.desc, objType, tc.isAccepted, response.Accepted)
				if response.Message != nil {
					t.Errorf(
						"[%s/%s] policy message: %s",
						tc.desc, objType, *response.Message)
				}
			}
			if response.MutatedObject == nil && tc.isMutated {
				t.Errorf("[%s/%s] object has not been mutated", tc.desc, objType)
			}
			if response.MutatedObject != nil && !tc.isMutated {
				t.Errorf("[%s/%s] object should not have been mutated", tc.desc, objType)
			}
		}
	}
}

func buildNamespaceJSON(annotations map[string]string) []byte {
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: annotations,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "namespace",
		},
	}
	namespaceJSON, err := json.Marshal(&namespace)
	if err != nil {
		log.Fatalf("cannot marshall namespace: %v", err)
	}

	return namespaceJSON
}

func buildServiceJSON(annotations map[string]string) []byte {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Namespace:   "default",
			Annotations: annotations,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "service",
		},
	}
	serviceSON, err := json.Marshal(&service)
	if err != nil {
		log.Fatalf("cannot marshall namespace: %v", err)
	}

	return serviceSON
}