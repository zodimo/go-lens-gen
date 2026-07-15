## 1. Walker & Model Changes

- [x] 1.1 Add `ArrayBasePath string` to `LensField` struct in `internal/gen/template.go`
- [x] 1.2 Update walker (`internal/gen/walker.go`) to populate `ArrayBasePath` on leaf fields discovered inside arrays
- [x] 1.3 Add `IsArrayItem bool` to `TemplateField` struct in `internal/gen/template.go`
- [x] 1.4 Update `prepareTemplateData` to set `IsArrayItem` when `ArrayBasePath` is non-empty
- [x] 1.5 Verify walker still handles `items` as `*jsonschema.Schema` (skip boolean `items` with debug log)

## 2. Template Engine Rewrite

- [x] 2.1 Replace `dynamicKeyRegex` with a unified token parser in `internal/gen/template.go`
- [x] 2.2 Implement `tokenizePath(path string) []pathToken` that scans left-to-right and emits tokens for:
  - literal segments
  - `<DYNAMIC_KEY_parent_N>` → `%s` + `parentKeyN string`
  - `<INDEX>` → `%d` + `indexN int`
- [x] 2.3 Update `prepareTemplateData` to use the new tokenizer for `FmtPath`, `MethodArgs`, `ArgsList`
- [x] 2.4 Update `prepareTemplateData` to append `At` suffix to `MethodName` when `IsArrayItem` is true
- [x] 2.5 Add unit tests for `tokenizePath` covering:
  - pure static path (no tokens)
  - pure dynamic key path
  - pure array index path
  - mixed dynamic + array (both orderings)
  - nested arrays (two `<INDEX>` tokens)

## 3. Code Generation Templates

- [x] 3.1 Add `Len` method template to `internal/gen/lens_template.go`
- [x] 3.2 Add `ForEach` method template to `internal/gen/lens_template.go`
- [x] 3.3 Update `TemplateContext` in `internal/gen/template.go` if new global fields are needed
- [x] 3.4 Verify generated code compiles with `gofmt` for a sample array schema

## 4. Integration & Wiring

- [x] 4.1 Update `GenerateCode` in `internal/gen/generate_code.go` to pass new template fields
- [x] 4.2 Run the CLI against `movie.schema.json` and inspect generated output
- [x] 4.3 Run the CLI against `blog-post.schema.json` and inspect generated output
- [x] 4.4 Run the CLI against `database.schema.json` (complex nested + dynamic) and verify no regressions

## 5. Test Coverage

- [x] 5.1 Create `internal/gen/template_test.go` with table-driven tests for `prepareTemplateData`
- [x] 5.2 Create `internal/gen/walker_test.go` with tests for array path discovery
- [x] 5.3 Add end-to-end test: compile `movie.schema.json` → generate → verify `GetCastAt`, `SetCastAt`, `LenCast`, `ForEachCast` exist
- [x] 5.4 Add end-to-end test for `blog-post.schema.json` → verify `GetTagsAt`, `SetTagsAt`, `LenTags`, `ForEachTags` exist
- [x] 5.5 Add integration test with real JSON payload: create lens, set array item, get array item, verify length, iterate with ForEach
- [x] 5.6 Run all existing tests (`go test ./...`) and verify no regressions

## 6. Documentation & Schema Updates

- [x] 6.1 Update `docs/lens-gen.md` Part 3 — mark `type: array` and `items` as ✅ Supported
- [x] 6.2 Update `docs/lens-gen.md` Part 3 test schema table — add method counts for array schemas
- [x] 6.3 Update `docs/lens-gen.md` Phase 2 roadmap — mark items 6 and 7 as complete
- [x] 6.4 Add note to `docs/lens-gen.md` about `items: true/false` being skipped

## 7. Validation

- [x] 7.1 Run `go test ./...` — all tests pass
- [x] 7.2 Run `go build ./...` — no compilation errors
- [x] 7.3 Run `go vet ./...` — no warnings
- [x] 7.4 Regenerate all `scratch/` outputs and verify they compile
- [x] 7.5 Run `test_all_schemas.sh` (if it exists) and confirm success
