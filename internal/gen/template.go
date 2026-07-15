package gen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type TemplateContext struct {
	PackageName string
	StructName  string
	Fields      []TemplateField
	Arrays      []TemplateArray
	IsDynamic   bool
}

type TemplateField struct {
	OriginalPath    string
	MethodName      string
	BaseMethodName  string
	MethodArgs      string
	ArgsList        string
	FmtPath         string
	GoType          string
	GjsonMethod     string
	IsDynamic       bool
	IsArrayItem     bool
	ArrayBasePath   string
	GenerateForEach bool
}

type TemplateArray struct {
	OriginalPath    string
	MethodName      string
	FmtPath         string
	MethodArgs      string
	ArgsList        string
	IsDynamic       bool
	GenerateForEach bool
	LeafGoType      string
	LeafGjsonMethod string
}

type LensField struct {
	Path          string
	GoType        string
	IsDynamic     bool
	ArrayBasePath string
}

var dynamicKeyRegex = regexp.MustCompile(`<DYNAMIC_KEY_([a-zA-Z0-9]+)_(\d+)>`)

type pathToken struct {
	IsDynamic bool
	IsIndex   bool
	Literal   string
	ArgName   string
	ArgType   string
	FmtPart   string
}

func tokenizePath(path string) []pathToken {
	var tokens []pathToken
	segments := strings.Split(path, ".")
	indexCount := 0

	for _, seg := range segments {
		if seg == "<INDEX>" {
			tokens = append(tokens, pathToken{
				IsIndex: true,
				ArgName: fmt.Sprintf("index%d", indexCount),
				ArgType: "int",
				FmtPart: "%d",
			})
			indexCount++
		} else if matches := dynamicKeyRegex.FindStringSubmatch(seg); len(matches) > 0 {
			tokens = append(tokens, pathToken{
				IsDynamic: true,
				ArgName:   fmt.Sprintf("%sKey%s", matches[1], matches[2]),
				ArgType:   "string",
				FmtPart:   "%s",
			})
		} else {
			tokens = append(tokens, pathToken{
				Literal: seg,
				FmtPart: seg,
			})
		}
	}
	return tokens
}

func prepareTemplateData(rawFields map[string]LensField) ([]TemplateField, []TemplateArray) {
	var tplFields []TemplateField
	arrayMap := make(map[string]bool)

	for path, field := range rawFields {
		tf := TemplateField{
			OriginalPath:  path,
			GoType:        field.GoType,
			IsDynamic:     field.IsDynamic,
			ArrayBasePath: field.ArrayBasePath,
			IsArrayItem:   field.ArrayBasePath != "",
		}

		if field.ArrayBasePath != "" {
			arrayMap[field.ArrayBasePath] = true
		}

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

		tokens := tokenizePath(path)
		var argsDef []string
		var argsCall []string
		var fmtParts []string
		var cleanSegments []string

		for _, t := range tokens {
			fmtParts = append(fmtParts, t.FmtPart)
			if t.IsDynamic || t.IsIndex {
				argsDef = append(argsDef, fmt.Sprintf("%s %s", t.ArgName, t.ArgType))
				argsCall = append(argsCall, t.ArgName)
			} else {
				cleanSegments = append(cleanSegments, strings.Title(t.Literal))
			}
		}

		tf.FmtPath = strings.Join(fmtParts, ".")
		tf.BaseMethodName = strings.Join(cleanSegments, "")
		tf.MethodName = tf.BaseMethodName
		if tf.IsArrayItem {
			tf.MethodName += "At"
		}
		tf.MethodArgs = strings.Join(argsDef, ", ")
		tf.ArgsList = strings.Join(argsCall, ", ")
		tf.GenerateForEach = strings.HasSuffix(path, ".<INDEX>")

		tplFields = append(tplFields, tf)
	}

	sort.Slice(tplFields, func(i, j int) bool {
		return tplFields[i].OriginalPath < tplFields[j].OriginalPath
	})

	var tplArrays []TemplateArray
	for arrPath := range arrayMap {
		ta := TemplateArray{
			OriginalPath: arrPath,
		}

		tokens := tokenizePath(arrPath)
		var argsDef []string
		var argsCall []string
		var fmtParts []string
		var cleanSegments []string

		for _, t := range tokens {
			fmtParts = append(fmtParts, t.FmtPart)
			if t.IsDynamic || t.IsIndex {
				argsDef = append(argsDef, fmt.Sprintf("%s %s", t.ArgName, t.ArgType))
				argsCall = append(argsCall, t.ArgName)
				ta.IsDynamic = true
			} else {
				cleanSegments = append(cleanSegments, strings.Title(t.Literal))
			}
		}

		ta.FmtPath = strings.Join(fmtParts, ".")
		ta.MethodName = strings.Join(cleanSegments, "")
		ta.MethodArgs = strings.Join(argsDef, ", ")
		ta.ArgsList = strings.Join(argsCall, ", ")

		// Check if it's a primitive array for ForEach
		for _, tf := range tplFields {
			if tf.OriginalPath == arrPath+".<INDEX>" {
				ta.GenerateForEach = true
				ta.LeafGoType = tf.GoType
				ta.LeafGjsonMethod = tf.GjsonMethod
				break
			}
		}

		tplArrays = append(tplArrays, ta)
	}

	sort.Slice(tplArrays, func(i, j int) bool {
		return tplArrays[i].OriginalPath < tplArrays[j].OriginalPath
	})

	return tplFields, tplArrays
}
