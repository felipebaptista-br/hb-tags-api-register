package visibility

import "encoding/json"

// PublicProfile returns the nested subset of profileData that is public.
// Visibility for each fields_schema key comes from profileData.visibility when
// that override is public, restricted, or private; otherwise it comes from the
// field's visibility in the category schema. Restricted, private, and unknown
// keys are omitted. An unreadable schema or profile yields an empty object.
func PublicProfile(profileData, fieldsSchema []byte) map[string]any {
	out := map[string]any{}
	fields, ok := schemaFields(fieldsSchema)
	if !ok {
		return out
	}
	profile, ok := decodeObject(profileData)
	if !ok {
		return out
	}

	overrides := map[string]string{}
	if raw, exists := profile["visibility"]; exists {
		if vis, isMap := raw.(map[string]any); isMap {
			for key, value := range vis {
				text, isText := value.(string)
				if isText && knownVisibility(text) {
					overrides[key] = text
				}
			}
		}
		delete(profile, "visibility")
	}

	for _, field := range fields {
		if field.Key == "" || field.Key == "visibility" {
			continue
		}
		vis := field.Visibility
		if override, exists := overrides[field.Key]; exists {
			vis = override
		}
		if vis != "public" {
			continue
		}
		value, exists := valueAt(profile, field.Key)
		if !exists || value == nil {
			continue
		}
		setPath(out, field.Key, value)
	}
	return out
}

type fieldDef struct {
	Key        string `json:"key"`
	Visibility string `json:"visibility"`
}

func schemaFields(raw []byte) ([]fieldDef, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, false
	}
	var schema struct {
		Fields []fieldDef `json:"fields"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, false
	}
	if schema.Fields == nil {
		return nil, false
	}
	return schema.Fields, true
}

func decodeObject(raw []byte) (map[string]any, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}, true
	}
	var profile map[string]any
	if err := json.Unmarshal(raw, &profile); err != nil || profile == nil {
		return nil, false
	}
	return profile, true
}

func knownVisibility(value string) bool {
	return value == "public" || value == "restricted" || value == "private"
}

func valueAt(root map[string]any, key string) (any, bool) {
	current := any(root)
	for _, part := range splitKey(key) {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, exists := object[part]
		if !exists {
			return nil, false
		}
		current = next
	}
	return current, true
}

func setPath(dst map[string]any, key string, value any) {
	parts := splitKey(key)
	current := dst
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = value
}

func splitKey(key string) []string {
	parts := make([]string, 0, 4)
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	return append(parts, key[start:])
}
