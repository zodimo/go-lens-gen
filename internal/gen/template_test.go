package gen

import (
	"reflect"
	"testing"
)

func TestTokenizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []pathToken
	}{
		{
			name: "pure static path",
			path: "users.name",
			expected: []pathToken{
				{Literal: "users", FmtPart: "users"},
				{Literal: "name", FmtPart: "name"},
			},
		},
		{
			name: "pure dynamic key path",
			path: "users.<DYNAMIC_KEY_users_0>.name",
			expected: []pathToken{
				{Literal: "users", FmtPart: "users"},
				{IsDynamic: true, ArgName: "usersKey0", ArgType: "string", FmtPart: "%s"},
				{Literal: "name", FmtPart: "name"},
			},
		},
		{
			name: "pure array index path",
			path: "cast.<INDEX>.name",
			expected: []pathToken{
				{Literal: "cast", FmtPart: "cast"},
				{IsIndex: true, ArgName: "index0", ArgType: "int", FmtPart: "%d"},
				{Literal: "name", FmtPart: "name"},
			},
		},
		{
			name: "mixed dynamic + array",
			path: "users.<DYNAMIC_KEY_users_0>.tags.<INDEX>",
			expected: []pathToken{
				{Literal: "users", FmtPart: "users"},
				{IsDynamic: true, ArgName: "usersKey0", ArgType: "string", FmtPart: "%s"},
				{Literal: "tags", FmtPart: "tags"},
				{IsIndex: true, ArgName: "index0", ArgType: "int", FmtPart: "%d"},
			},
		},
		{
			name: "nested arrays",
			path: "matrix.<INDEX>.<INDEX>",
			expected: []pathToken{
				{Literal: "matrix", FmtPart: "matrix"},
				{IsIndex: true, ArgName: "index0", ArgType: "int", FmtPart: "%d"},
				{IsIndex: true, ArgName: "index1", ArgType: "int", FmtPart: "%d"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tokenizePath(tc.path)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestPrepareTemplateData(t *testing.T) {
	rawFields := map[string]LensField{
		"users.<DYNAMIC_KEY_users_0>.tags.<INDEX>": {
			Path:          "users.<DYNAMIC_KEY_users_0>.tags.<INDEX>",
			GoType:        "string",
			IsDynamic:     true,
			ArrayBasePath: "users.<DYNAMIC_KEY_users_0>.tags",
		},
		"cast.<INDEX>.name": {
			Path:          "cast.<INDEX>.name",
			GoType:        "string",
			IsDynamic:     true,
			ArrayBasePath: "cast",
		},
	}

	fields, arrays := prepareTemplateData(rawFields)

	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
	if len(arrays) != 2 {
		t.Errorf("expected 2 arrays, got %d", len(arrays))
	}

	// arrays should be sorted
	if arrays[0].OriginalPath != "cast" {
		t.Errorf("expected first array to be cast, got %s", arrays[0].OriginalPath)
	}
	if arrays[1].OriginalPath != "users.<DYNAMIC_KEY_users_0>.tags" {
		t.Errorf("expected second array to be users...tags, got %s", arrays[1].OriginalPath)
	}

	if arrays[1].MethodArgs != "usersKey0 string" {
		t.Errorf("expected usersKey0 string, got %s", arrays[1].MethodArgs)
	}
	if !arrays[1].GenerateForEach {
		t.Errorf("expected GenerateForEach to be true for tags array")
	}
	if arrays[0].GenerateForEach {
		t.Errorf("expected GenerateForEach to be false for cast array (it has object items)")
	}

	for _, f := range fields {
		if f.OriginalPath == "cast.<INDEX>.name" {
			if f.MethodName != "CastNameAt" {
				t.Errorf("expected CastNameAt, got %s", f.MethodName)
			}
			if f.MethodArgs != "index0 int" {
				t.Errorf("expected index0 int, got %s", f.MethodArgs)
			}
		}
	}
}
