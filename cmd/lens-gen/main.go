package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zodimo/go-lens-gen/internal/gen"
)

func main() {

	// 1. Define CLI Flags for dynamic configuration
	pkgName := flag.String("pkg", "", "The package name for the generated file")
	lensName := flag.String("struct", "", "The name of the generated struct")
	schemaPath := flag.String("schema", "", "Path or URL to the input JSON schema")
	outPath := flag.String("out", "", "Output path for the generated code")

	logger := slog.Default()

	flag.Parse()

	if *pkgName == "" {
		log.Fatalf("pkg is required")
	}
	if *lensName == "" {
		log.Fatalf("struct is required")
	}
	if *schemaPath == "" {
		log.Fatalf("schema is required")
	}

	if *outPath == "" {
		// should we fail or write to stdout ?
		log.Fatalf("out is required")
	}

	logger.Info("Starting Lens Generator...")
	logger.Info(fmt.Sprintf("Package: %s | Struct: %s", *pkgName, *lensName))

	// 2. Ensure the output directory exists
	if err := os.MkdirAll(filepath.Dir(*outPath), 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	generator, err := gen.NewGenerator(
		*schemaPath,
		*pkgName,
		*lensName,
		gen.WithLogger(logger.With("compoenent", "gen.Generator")),
	)

	if err != nil {
		log.Fatalf("Failed to Create Generator: %v", err)
	}

	// // 3. Compile the JSON Schema using the library
	// // This automatically validates the schema and resolves any $ref pointers internally
	// compiler := jsonschema.NewCompiler()

	// // We read the file from disk using the compiler's built-in loader
	// sch, err := compiler.Compile(*schemaPath)
	// if err != nil {
	// 	log.Fatalf("Failed to compile JSON schema: %v", err)
	// }

	// // 4. Initialize our map and start the recursive tree-walker
	// fields := make(map[string]gen.LensField)
	// log.Println("Walking schema tree to discover paths and types...")
	// gen.WalkSchema(sch, "", fields)

	// if len(fields) == 0 {
	// 	log.Fatal("Walker found 0 fields. Is your schema valid and populated?")
	// }
	// log.Printf("Discovered %d unique paths.", len(fields))

	// 5. Write to file
	fileHandle, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("Failed to Creating File: %v", err)
	}
	defer fileHandle.Close()

	// // 5. Pass the extracted data into the code generator
	// log.Println("Code generation complete! Your type-safe lenses are ready to use.")

	// log.Printf("Generating Go code at: %s", *outPath)
	// gen.GenerateCode(fields, *outPath, *pkgName, *structName)

	// log.Println("Code generation complete!")

	n, err := generator.Generate(fileHandle)
	if err != nil {
		log.Fatalf("Failed to Generate: %v", err)
	}
	if n == 0 {
		// something went wrong...
		log.Fatalf("Failed to Generate: %v", err)
	}

	fmt.Printf("Successfully generated lenses in %s\n", *outPath)
}
