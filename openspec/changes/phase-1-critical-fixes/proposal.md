## Why

Lens-Gen currently fails on 8 out of 10 real-world test schemas from json-schema.org. The generator crashes with `gofmt` errors when package names contain hyphens, cannot resolve external `$ref` pointers, and silently ignores internal `$defs` and `$ref` anchors. These are not edge cases — they are blocking bugs that prevent the tool from handling any non-trivial JSON Schema. Fixing them is prerequisite to all other features.

## What Changes

- **Sanitize `--pkg` values** to produce valid Go package names (strip/replace hyphens and invalid characters).
- **Sanitize `--struct` values** to produce valid Go identifiers.
- **Add a custom JSON Schema loader** that resolves external `$ref` URLs to local filesystem paths, enabling schemas with cross-file references to compile.
- **Traverse internal `$ref` and `$defs`/`definitions`** in the tree walker so that referenced schemas produce lens methods instead of being silently skipped.

## Capabilities

### New Capabilities

- `package-sanitization`: Sanitize CLI `--pkg` and `--struct` inputs into valid Go identifiers.
- `external-ref-loader`: Resolve external JSON Schema `$ref` pointers via a custom filesystem-based loader.
- `internal-ref-traversal`: Recursively follow internal `$ref`, `$defs`, and `definitions` during tree walking.

### Modified Capabilities

- (none — this is purely fixing existing behavior, not changing requirements)

## Impact

- `cmd/lens-gen/main.go`: CLI flag validation will delegate sanitization to a helper.
- `internal/gen/generator.go`: Generator construction will register a custom loader with the `jsonschema` compiler.
- `internal/gen/walker.go`: `WalkSchema` will detect `$ref` pointers, resolve them, and recurse into the referenced schema.
- All 10 `testdata/schemas/*.json` files will become usable for code generation and regression testing.
