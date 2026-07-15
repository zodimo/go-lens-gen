package templates

import (
	"embed"
	"fmt"
	"text/template"
)

//go:embed lens/*.go.tpl
var templateFS embed.FS

var Templates *template.Template

func init() {
	var err error
	Templates, err = template.ParseFS(templateFS, "lens/*.go.tpl")
	if err != nil {
		panic(fmt.Errorf("Could not load the templates: %w", err))
	}
	// lens.go.tpl is the main template; it references the "header" template.
	Templates = Templates.Lookup("lens.go.tpl")
	if Templates == nil {
		panic("main template lens.go.tpl not found")
	}
}
