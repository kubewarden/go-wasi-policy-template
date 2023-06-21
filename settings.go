package main

import (
	"encoding/json"
	"fmt"
	"log"

	mapset "github.com/deckarep/golang-set/v2"
)

// Settings defines the settings of the policy
type Settings struct {
	RequiredAnnotations  map[string]string  `json:"requiredAnnotations"`
	ForbiddenAnnotations mapset.Set[string] `json:"forbiddenAnnotations"`
}

func validateSettings(input []byte) []byte {
	var response SettingsValidationResponse

	settings := &Settings{
		// this is required to make the unmarshal work
		ForbiddenAnnotations: mapset.NewSet[string](),
	}
	if err := json.Unmarshal(input, &settings); err != nil {
		response = RejectSettings(Message(fmt.Sprintf("cannot unmarshal settings: %v", err)))
	} else {
		response = validateCliSettings(settings)
	}

	responseBytes, err := json.Marshal(&response)
	if err != nil {
		log.Fatalf("cannot marshal validation response: %v", err)
	}
	return responseBytes
}

func validateCliSettings(settings *Settings) SettingsValidationResponse {
	required := mapset.NewSet[string]()
	for key := range settings.RequiredAnnotations {
		required.Add(key)
	}

	forbiddenButRequired := settings.ForbiddenAnnotations.Intersect(required)

	if forbiddenButRequired.Cardinality() > 0 {
		return RejectSettings(Message(
			fmt.Sprintf("The following annotations are forbidden and required at the same time: %s", forbiddenButRequired.String())))
	}

	return AcceptSettings()
}
