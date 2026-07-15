package gen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// TemplateContext holds the global generation variables and the parsed fields
type TemplateContext struct {
	PackageName string
	StructName  string
	Fields      []TemplateField
	IsDynamic   bool
}

// TemplateField holds the enriched data for a single getter/setter pair
type TemplateField struct {
	OriginalPath string
	MethodName   string // e.g., OrganizationsUsersAge
	MethodArgs   string // e.g., organizationsKey0 string, usersKey1 string
	ArgsList     string // e.g., organizationsKey0, usersKey1
	FmtPath      string // e.g., organizations.%s.users.%s.age
	GoType       string // e.g., int64
	GjsonMethod  string // e.g., Int
	IsDynamic    bool
}

// LensField holds the metadata required for the text/template to generate code
type LensField struct {
	Path      string // The exact dot-notation path (e.g., "features.beta_ui" or "organizations.<DYNAMIC_KEY_organizations_0>.name")
	GoType    string // The inferred Go primitive type (e.g., "string", "int64", "bool")
	IsDynamic bool   // A flag indicating if the path requires runtime arguments (contains a dynamic placeholder)
}

// Regex to find tokens like <DYNAMIC_KEY_users_1>
var dynamicKeyRegex = regexp.MustCompile(`<DYNAMIC_KEY_([a-zA-Z0-9]+)_(\d+)>`)

// prepareTemplateData transforms the walker's raw map into structured template fields
func prepareTemplateData(rawFields map[string]LensField) []TemplateField {
	var tplFields []TemplateField

	for path, field := range rawFields {
		tf := TemplateField{
			OriginalPath: path,
			GoType:       field.GoType,
			IsDynamic:    field.IsDynamic,
		}

		// Map Go types to gjson extraction methods
		switch field.GoType {
		case "int64":
			tf.GjsonMethod = "Int"
		case "float64":
			tf.GjsonMethod = "Float"
		case "bool":
			tf.GjsonMethod = "Bool"
		default:
			tf.GjsonMethod = "String"
		}

		// Parse Dynamic Keys
		matches := dynamicKeyRegex.FindAllStringSubmatch(path, -1)

		var argsDef []string
		var argsCall []string
		cleanPathSegments := []string{}

		// Split path by dots to build the PascalCase method name
		rawSegments := strings.Split(path, ".")
		for _, seg := range rawSegments {
			if dynamicKeyRegex.MatchString(seg) {
				continue // Skip the dynamic tokens in the method name
			}
			cleanPathSegments = append(cleanPathSegments, strings.Title(seg))
		}
		tf.MethodName = strings.Join(cleanPathSegments, "")

		// Build the argument lists and the fmt.Sprintf path
		tf.FmtPath = dynamicKeyRegex.ReplaceAllString(path, "%s")

		for _, match := range matches {
			// match[1] is the parent name (e.g., users), match[2] is the index
			argName := fmt.Sprintf("%sKey%s", match[1], match[2])
			argsDef = append(argsDef, fmt.Sprintf("%s string", argName))
			argsCall = append(argsCall, argName)
		}

		tf.MethodArgs = strings.Join(argsDef, ", ")
		tf.ArgsList = strings.Join(argsCall, ", ")

		tplFields = append(tplFields, tf)
	}

	// Sort alphabetically by path for deterministic code generation
	sort.Slice(tplFields, func(i, j int) bool {
		return tplFields[i].OriginalPath < tplFields[j].OriginalPath
	})

	return tplFields
}
