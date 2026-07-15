package gen

import (
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestWalkSchema_Arrays(t *testing.T) {
	schemaJSON := `{
		"$id": "https://example.com/test.schema.json",
		"type": "object",
		"properties": {
			"tags": {
				"type": "array",
				"items": { "type": "string" }
			},
			"matrix": {
				"type": "array",
				"items": {
					"type": "array",
					"items": { "type": "integer" }
				}
			},
			"skippedBool": {
				"type": "array",
				"items": true
			}
		},
		"additionalProperties": {
			"type": "object",
			"properties": {
				"history": {
					"type": "array",
					"items": { "type": "string" }
				}
			}
		}
	}`

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("test.schema.json", strings.NewReader(schemaJSON)); err != nil {
		t.Fatalf("Failed to add resource: %v", err)
	}

	sch, err := compiler.Compile("test.schema.json")
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	fields := make(map[string]LensField)
	WalkSchema(sch, "", fields)

	expectedPaths := map[string]struct{
		goType string
		arrayBasePath string
	}{
		"tags.<INDEX>": {goType: "string", arrayBasePath: "tags"},
		"matrix.<INDEX>.<INDEX>": {goType: "int64", arrayBasePath: "matrix.<INDEX>"},
		"<DYNAMIC_KEY_root_0>.history.<INDEX>": {goType: "string", arrayBasePath: "<DYNAMIC_KEY_root_0>.history"},
	}

	if len(fields) != len(expectedPaths) {
		t.Errorf("expected %d fields, got %d", len(expectedPaths), len(fields))
	}

	for path, expected := range expectedPaths {
		field, ok := fields[path]
		if !ok {
			t.Errorf("missing expected path: %s", path)
			continue
		}
		if field.GoType != expected.goType {
			t.Errorf("expected path %s to have GoType %s, got %s", path, expected.goType, field.GoType)
		}
		if field.ArrayBasePath != expected.arrayBasePath {
			t.Errorf("expected path %s to have ArrayBasePath %s, got %s", path, expected.arrayBasePath, field.ArrayBasePath)
		}
	}
}
