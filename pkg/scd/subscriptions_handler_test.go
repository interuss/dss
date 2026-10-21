package scd

import (
	"testing"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
)

func boolPointer(value bool) *bool {
	return &value
}

func TestValidateRequestedSubscriptionScopes(t *testing.T) {
	tests := []struct {
		name        string
		scopes      []string
		opIntents   bool
		constraints bool
		wantError   bool
	}{
		{
			name:      "strategic permits operational intents",
			scopes:    []string{string(restapi.UtmStrategicCoordinationScope)},
			opIntents: true,
		},
		{
			name:        "constraint processing permits constraints",
			scopes:      []string{string(restapi.UtmConstraintProcessingScope)},
			constraints: true,
		},
		{
			name:        "both scopes permit both notification types",
			scopes:      []string{string(restapi.UtmStrategicCoordinationScope), string(restapi.UtmConstraintProcessingScope)},
			opIntents:   true,
			constraints: true,
		},
		{
			name:        "strategic does not permit constraints",
			scopes:      []string{string(restapi.UtmStrategicCoordinationScope)},
			constraints: true,
			wantError:   true,
		},
		{
			name:      "constraint processing does not permit operational intents",
			scopes:    []string{string(restapi.UtmConstraintProcessingScope)},
			opIntents: true,
			wantError: true,
		},
		{
			name:        "strategic does not permit both notification types",
			scopes:      []string{string(restapi.UtmStrategicCoordinationScope)},
			opIntents:   true,
			constraints: true,
			wantError:   true,
		},
		{
			name:        "constraint processing does not permit both notification types",
			scopes:      []string{string(restapi.UtmConstraintProcessingScope)},
			opIntents:   true,
			constraints: true,
			wantError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRequestedSubscriptionScopes(
				api.AuthorizationResult{Scopes: test.scopes},
				&restapi.PutSubscriptionParameters{
					NotifyForOperationalIntents: boolPointer(test.opIntents),
					NotifyForConstraints:        boolPointer(test.constraints),
				},
			)
			if (err != nil) != test.wantError {
				t.Fatalf("validateRequestedSubscriptionScopes() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}
