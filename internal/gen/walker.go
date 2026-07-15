package gen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// WalkSchema recursively traverses the compiled JSON Schema
func WalkSchema(sch *jsonschema.Schema, currentPath string, fields map[string]LensField) {
	if sch == nil {
		return
	}

	if len(sch.Types) == 0 {
		jsonString, _ := json.MarshalIndent(sch, "", "  ")
		_ = jsonString
		// jsonString := ""
		// sch.OneOf
		// fmt.Printf("Nothing here.. :%s\n%s", currentPath, string(jsonString))

		if len(sch.OneOf) > 0 {
			fmt.Printf("OneOf at: %s\n", currentPath)
		}
		if len(sch.AnyOf) > 0 {
			fmt.Printf("AnyOf at: %s\n", currentPath)
			// strayegy - flatten

		}
		if len(sch.AllOf) > 0 {
			fmt.Printf("AllOf at: %s\n", currentPath)
			// strayegy - flatten

		}

		return
	}

	if len(sch.Types) > 1 {
		fmt.Printf("more the 1 type at path ??: %s, %v\n", currentPath, sch.Types)
	}

	primaryType := sch.Types[0]

	switch primaryType {
	case "object":
		// 1. Handle standard, static properties
		for key, propSchema := range sch.Properties {
			nextPath := key
			if currentPath != "" {
				nextPath = currentPath + "." + key
			}
			WalkSchema(propSchema, nextPath, fields)
		}

		// 2. Handle Dynamic Pattern Properties
		if patSchema, ok := sch.AdditionalProperties.(*jsonschema.Schema); ok {
			// 1. Identify the parent node's name
			parentName := "root"
			if currentPath != "" {
				segments := strings.Split(currentPath, ".")
				parentName = segments[len(segments)-1]
			}

			// 2. Count existing keys to prevent collisions if the schema repeats parent names
			dynamicDepth := strings.Count(currentPath, "<DYNAMIC_KEY")

			// 3. Construct the semantic placeholder (e.g., <DYNAMIC_KEY_users_0>)
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

		//3. dependencies

		for depKey, depValue := range sch.Dependencies {
			nextPath := depKey
			if currentPath != "" {
				nextPath = currentPath + "." + depKey
			}
			fmt.Printf("key:%s, depValue: %T\n", depKey, depValue)
			if depSchema, ok := depValue.(*jsonschema.Schema); ok {
				WalkSchema(depSchema, nextPath, fields)
			}

			// if depSchema, ok := depValue.(*jsonschema.); ok {
			// 	walkSchema(depSchema, nextPath, fields)
			// }

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
