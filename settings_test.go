package main

import (
	"encoding/json"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
)

func TestValidateSettings(t *testing.T) {
	cases := []struct {
		desc                 string
		requiredAnnotations  map[string]string
		forbiddenAnnotations mapset.Set[string]
		isValid              bool
	}{
		{
			"empty",
			map[string]string{},
			mapset.NewSet[string](),
			true,
		},
		{
			"only required annotations",
			map[string]string{
				"cc-center": "marketing",
			},
			mapset.NewSet[string](),
			true,
		},
		{
			"only forbidden annotations",
			map[string]string{},
			mapset.NewSet[string]("priority"),
			true,
		},
		{
			"no contradictions",
			map[string]string{
				"cc-center": "marketing",
			},
			mapset.NewSet[string]("priority"),
			true,
		},
		{
			"contradictions",
			map[string]string{
				"cc-center": "marketing",
			},
			mapset.NewSet[string]("cc-center"),
			false,
		},
	}

	for _, tc := range cases {
		settings := Settings{
			RequiredAnnotations:  tc.requiredAnnotations,
			ForbiddenAnnotations: tc.forbiddenAnnotations,
		}
		settingsJSON, err := json.Marshal(&settings)
		if err != nil {
			t.Errorf("[%s] cannot marshal settings: %v", tc.desc, err)
		}

		responseJSON := validateSettings(settingsJSON)
		var response SettingsValidationResponse
		err = json.Unmarshal(responseJSON, &response)
		if err != nil {
			t.Errorf("[%s] cannot unmarshal response: %v", tc.desc, err)
		}

		if response.Valid != tc.isValid {
			t.Errorf(
				"[%s] didn't get the expected validation outcome, %v was expected, got %v instead",
				tc.desc, tc.isValid, response.Valid)
			if response.Message != nil {
				t.Errorf(
					"[%s] validation message: %s",
					tc.desc, *response.Message)
			}
		}
	}
}
