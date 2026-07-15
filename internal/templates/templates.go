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
}
