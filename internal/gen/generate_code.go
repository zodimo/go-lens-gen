package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io"

	"github.com/zodimo/go-lens-gen/internal/templates"
)

func (g *Generator) GenerateCode(rawFields map[string]LensField, outputWriter io.Writer, pkgName string, structName string) (n int64, err error) {
	// Pre-process the data
	templateData, templateArrays := prepareTemplateData(rawFields)

	var isDynamic bool

	for _, templateField := range templateData {
		if templateField.IsDynamic {
			isDynamic = true
			break
		}
	}
	for _, templateArray := range templateArrays {
		if templateArray.IsDynamic {
			isDynamic = true
			break
		}
	}

	// Build the context
	ctx := TemplateContext{
		PackageName: pkgName,
		StructName:  structName,
		Fields:      templateData,
		Arrays:      templateArrays,
		IsDynamic:   isDynamic,
	}

	tmpl := templates.Templates

	var buf bytes.Buffer
	// Pass 'ctx' instead of 'templateData'
	if err = tmpl.Execute(&buf, ctx); err != nil {
		err = fmt.Errorf("template execution error: %w", err)
		g.options.logger.Error("template execution error", "error", err)
		return
	}

	var formattedCode []byte
	// Format the generated Go code (gofmt)
	formattedCode, err = format.Source(buf.Bytes())
	if err != nil {

		err = fmt.Errorf("gofmt error: %w", err)
		// If formatting fails, print the raw buffer so you can debug the syntax error
		g.options.logger.Debug("Formatter code", "code", buf.String())
		g.options.logger.Error("gofmt error", "error", err)

		return
	}

	return io.Copy(outputWriter, bytes.NewReader(formattedCode))

}
