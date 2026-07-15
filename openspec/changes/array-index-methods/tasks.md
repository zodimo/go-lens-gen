## 1. Walker & Model Changes

- [ ] 1.1 Add `ArrayBasePath string` to `LensField` struct in `internal/gen/template.go`
- [ ] 1.2 Update walker (`internal/gen/walker.go`) to populate `ArrayBasePath` on leaf fields discovered inside arrays
- [ ] 1.3 Add `IsArrayItem bool` to `TemplateField` struct in `internal/gen/template.go`
- [ ] 1.4 Update `prepareTemplateData` to set `IsArrayItem` when `ArrayBasePath` is non-empty
- [ ] 1.5 Verify walker still handles `items` as `*jsonschema.Schema` (skip boolean `items` with debug log)

## 2. Template Engine Rewrite

- [ ] 2.1 Replace `dynamicKeyRegex` with a unified token parser in `internal/gen/template.go`
- [ ] 2.2 Implement `tokenizePath(path string) []pathToken` that scans left-to-right and emits tokens for:
  - literal segments
  - `<DYNAMIC_KEY_parent_N>` → `%s` + `parentKeyN string`
  - `<INDEX>` → `%d` + `indexN int`
- [ ] 2.3 Update `prepareTemplateData` to use the new tokenizer for `FmtPath`, `MethodArgs`, `ArgsList`
- [ ] 2.4 Update `prepareTemplateData` to append `At` suffix to `MethodName` when `IsArrayItem` is true
- [ ] 2.5 Add unit tests for `tokenizePath` covering:
  - pure static path (no tokens)
  - pure dynamic key path
  - pure array index path
  - mixed dynamic + array (both orderings)
  - nested arrays (two `<INDEX>` tokens)

## 3. Code Generation Templates

- [ ] 3.1 Add `Len` method template to `internal/gen/lens_template.go`
- [ ] 3.2 Add `ForEach` method template to `internal/gen/lens_template.go`
- [ ] 3.3 Update `TemplateContext` in `internal/gen/template.go` if new global fields are needed
- [ ] 3.4 Verify generated code compiles with `gofmt` for a sample array schema

## 4. Integration & Wiring

- [ ] 4.1 Update `GenerateCode` in `internal/gen/generate_code.go` to pass new template fields
- [ ] 4.2 Run the CLI against `movie.schema.json` and inspect generated output
- [ ] 4.3 Run the CLI against `blog-post.schema.json` and inspect generated output
- [ ] 4.4 Run the CLI against `database.schema.json` (complex nested + dynamic) and verify no regressions

## 5. Test Coverage

- [ ] 5.1 Create `internal/gen/template_test.go` with table-driven tests for `prepareTemplateData`
- [ ] 5.2 Create `internal/gen/walker_test.go` with tests for array path discovery
- [ ] 5.3 Add end-to-end test: compile `movie.schema.json` → generate → verify `GetCastAt`, `SetCastAt`, `LenCast`, `ForEachCast` exist
- [ ] 5.4 Add end-to-end test for `blog-post.schema.json` → verify `GetTagsAt`, `SetTagsAt`, `LenTags`, `ForEachTags` exist
- [ ] 5.5 Add integration test with real JSON payload: create lens, set array item, get array item, verify length, iterate with ForEach
- [ ] 5.6 Run all existing tests (`go test ./...`) and verify no regressions

## 6. Documentation & Schema Updates

- [ ] 6.1 Update `docs/lens-gen.md` Part 3 — mark `type: array` and `items` as ✅ Supported
- [ ] 6.2 Update `docs/lens-gen.md` Part 3 test schema table — add method counts for array schemas
- [ ] 6.3 Update `docs/lens-gen.md` Phase 2 roadmap — mark items 6 and 7 as complete
- [ ] 6.4 Add note to `docs/lens-gen.md` about `items: true/false` being skipped

## 7. Validation

- [ ] 7.1 Run `go test ./...` — all tests pass
- [ ] 7.2 Run `go build ./...` — no compilation errors
- [ ] 7.3 Run `go vet ./...` — no warnings
- [ ] 7.4 Regenerate all `scratch/` outputs and verify they compile
- [ ] 7.5 Run `test_all_schemas.sh` (if it exists) and confirm success
