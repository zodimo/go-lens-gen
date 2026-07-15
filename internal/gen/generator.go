package gen

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type GeneratorOptions struct {
	logger *slog.Logger
}

type GeneratorOption = func(o *GeneratorOptions)

func WithLogger(logger *slog.Logger) func(o *GeneratorOptions) {
	return func(o *GeneratorOptions) {
		o.logger = logger
	}
}

type Generator struct {
	schemaPath  string
	packageName string
	lensName    string

	schema *jsonschema.Schema

	options *GeneratorOptions
}

func NewGenerator(
	schemaPath string,
	packageName string,
	lensName string,
	options ...GeneratorOption,
) (generator *Generator, err error) {

	opts := GeneratorOptions{
		logger: slog.Default(),
	}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	if schemaPath == "" {
		err = fmt.Errorf("schemaPath is required")
		return
	}

	// Compile the JSON Schema using the library
	// This automatically validates the schema and resolves any $ref pointers internally
	compiler := jsonschema.NewCompiler()

	sch, err := compiler.Compile(schemaPath)
	if err != nil {
		err = fmt.Errorf("Failed to compile JSON schema: %w", err)
		return
	}

	generator = &Generator{
		schemaPath:  schemaPath,
		packageName: packageName,
		lensName:    lensName,
		schema:      sch,
		options:     &opts,
	}

	return
}

func (g *Generator) Generate(outputWriter io.Writer) (n int64, err error) {

	// 4. Initialize our map and start the recursive tree-walker
	fields := make(map[string]LensField)
	g.options.logger.Info("Walking schema tree to discover paths and types...")
	WalkSchema(g.schema, "", fields)

	if len(fields) == 0 {
		err = fmt.Errorf("Walker found 0 fields. Is your schema valid and populated?")
		g.options.logger.Warn("Walker found 0 fields. Is your schema valid and populated?")
	}
	g.options.logger.Info(fmt.Sprintf("Discovered %d unique paths.", len(fields)))

	// 5. Pass the extracted data into the code generator
	g.options.logger.Info("Code generation complete! Your type-safe lenses are ready to use.")

	n, err = g.GenerateCode(fields, outputWriter, g.packageName, g.lensName)

	if err == nil {
		g.options.logger.Info("Code generation complete!")
	}
	return
}
