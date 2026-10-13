package visibility

import (
	"encoding/json"
	"reflect"
	"testing"
)

const testSchema = `{
  "version": 1,
  "fields": [
    {"key": "identity.display_name", "visibility": "public"},
    {"key": "extra.microchip", "visibility": "restricted"},
    {"key": "extra.tutor_address", "visibility": "private"},
    {"key": "health.blood_type", "visibility": "public"},
    {"key": "health.health_plan.card_number", "visibility": "private"},
    {"key": "contacts", "visibility": "public"},
    {"key": "extra.diagnosis", "visibility": "private"},
    {"key": "extra.neutered", "visibility": "public"}
  ]
}`

func TestPublicProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		schema  string
		want    string
	}{
		{
			name: "public fields only",
			profile: `{
				"identity": {"display_name": "Rex"},
				"extra": {"microchip": "999", "tutor_address": "Rua 1", "diagnosis": "secret", "neutered": false},
				"health": {"blood_type": "O+", "health_plan": {"card_number": "123"}},
				"contacts": [{"name": "Ana"}],
				"unknown": "nope",
				"visibility": {"extra.microchip": "restricted"}
			}`,
			schema: testSchema,
			want: `{
				"identity": {"display_name": "Rex"},
				"extra": {"neutered": false},
				"health": {"blood_type": "O+"},
				"contacts": [{"name": "Ana"}]
			}`,
		},
		{
			name: "override public and private",
			profile: `{
				"identity": {"display_name": "Ana"},
				"extra": {"diagnosis": "Alzheimer"},
				"health": {"blood_type": "A+"},
				"visibility": {"health.blood_type": "private", "extra.diagnosis": "public"}
			}`,
			schema: testSchema,
			want: `{
				"identity": {"display_name": "Ana"},
				"extra": {"diagnosis": "Alzheimer"}
			}`,
		},
		{
			name: "unknown override keeps schema visibility",
			profile: `{
				"health": {"blood_type": "O-"},
				"visibility": {"health.blood_type": "PUBLIC"}
			}`,
			schema: testSchema,
			want:   `{"health": {"blood_type": "O-"}}`,
		},
		{
			name: "null and missing values are omitted",
			profile: `{
				"identity": {"display_name": "Rex"},
				"health": {"blood_type": null}
			}`,
			schema: testSchema,
			want:   `{"identity": {"display_name": "Rex"}}`,
		},
		{
			name:    "invalid schema yields an empty profile",
			profile: `{"identity": {"display_name": "Rex"}}`,
			schema:  `[]`,
			want:    `{}`,
		},
		{
			name:    "invalid profile yields an empty profile",
			profile: `["nope"]`,
			schema:  testSchema,
			want:    `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PublicProfile([]byte(tt.profile), []byte(tt.schema))
			if _, leaked := got["visibility"]; leaked {
				t.Fatal("visibility map was returned")
			}
			var want map[string]any
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatalf("want json: %v", err)
			}
			if want == nil {
				want = map[string]any{}
			}
			if !reflect.DeepEqual(got, want) {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(want)
				t.Fatalf("profile = %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}
