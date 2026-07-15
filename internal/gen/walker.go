package gen

import (
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// WalkSchema recursively traverses the compiled JSON Schema
func WalkSchema(sch *jsonschema.Schema, currentPath string, fields map[string]LensField) {
	if sch == nil {
		return
	}

	if sch.Ref != nil {
		if sch.Ref != sch {
			WalkSchema(sch.Ref, currentPath, fields)
		}
		return
	}

	if len(sch.OneOf) > 0 {
		for _, sub := range sch.OneOf {
			WalkSchema(sub, currentPath, fields)
		}
	}
	if len(sch.AnyOf) > 0 {
		for _, sub := range sch.AnyOf {
			WalkSchema(sub, currentPath, fields)
		}
	}
	if len(sch.AllOf) > 0 {
		for _, sub := range sch.AllOf {
			WalkSchema(sub, currentPath, fields)
		}
	}

	var primaryType string
	if len(sch.Types) == 0 {
		if len(sch.Properties) > 0 || sch.AdditionalProperties != nil || len(sch.PatternProperties) > 0 {
			primaryType = "object"
		} else if sch.Items != nil {
			primaryType = "array"
		} else {
			return
		}
	} else {
		if len(sch.Types) > 1 {
			fmt.Printf("more the 1 type at path ??: %s, %v\n", currentPath, sch.Types)
		}
		primaryType = sch.Types[0]
	}

	switch primaryType {
	case "object":
		// Handle standard, static properties
		for key, propSchema := range sch.Properties {
			nextPath := key
			if currentPath != "" {
				nextPath = currentPath + "." + key
			}
			WalkSchema(propSchema, nextPath, fields)
		}

		// Handle Dynamic Pattern Properties
		if patSchema, ok := sch.AdditionalProperties.(*jsonschema.Schema); ok {
			// Identify the parent node's name
			parentName := "root"
			if currentPath != "" {
				segments := strings.Split(currentPath, ".")
				parentName = segments[len(segments)-1]
			}

			// Count existing keys to prevent collisions if the schema repeats parent names
			dynamicDepth := strings.Count(currentPath, "<DYNAMIC_KEY")

			// Construct the semantic placeholder (e.g., <DYNAMIC_KEY_users_0>)
			placeholder := fmt.Sprintf("<DYNAMIC_KEY_%s_%d>", parentName, dynamicDepth)

			nextPath := currentPath + "." + placeholder
			if currentPath == "" {
				nextPath = placeholder
			}

			// Pass the semantic path down the tree
			WalkSchema(patSchema, nextPath, fields)

			// Note: We don't save the object itself as a field, we just recurse through it
			// to find the primitive leaf nodes (strings, ints, bools).

		}

		//dependencies

		for depKey, depValue := range sch.Dependencies {
			fmt.Printf("key:%s, depValue: %T\n", depKey, depValue)
			if depSchema, ok := depValue.(*jsonschema.Schema); ok {
				// Schema dependencies define additional properties at the SAME
				// object level (siblings), not as children of depKey.
				WalkSchema(depSchema, currentPath, fields)
			}
		}

	case "array":
		// For arrays, gjson uses '#' to indicate an array element or iteration.
		// We use <INDEX> as our placeholder.
		if sch.Items != nil {
			if itemSchema, ok := sch.Items.(*jsonschema.Schema); ok {
				nextPath := currentPath + ".<INDEX>"
				WalkSchema(itemSchema, nextPath, fields)
			}
		}

	// 3. Handle Primitive Leaf Nodes (This is what we actually generate getters/setters for)
	case "string", "integer", "number", "boolean":
		goType := "string"
		if primaryType == "integer" {
			goType = "int64"
		} else if primaryType == "number" {
			goType = "float64"
		} else if primaryType == "boolean" {
			goType = "bool"
		}

		fields[currentPath] = LensField{
			Path:      currentPath,
			GoType:    goType,
			IsDynamic: strings.Contains(currentPath, "<"),
		}
	default:
		panic(fmt.Sprintf("unhandled type: %s", primaryType))
	}
}
