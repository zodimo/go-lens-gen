## 1. Package & Struct Sanitization

- [x] 1.1 Create `internal/gen/sanitize.go` with `ValidatePackageName`, `ValidateStructName`, `SanitizePackageName`, and `SanitizeStructName`
- [x] 1.2 `ValidatePackageName` and `ValidateStructName`: return descriptive error if name is not valid Go identifier
- [x] 1.3 `SanitizePackageName`: strip all non `[a-zA-Z0-9_]` chars, lowercase, strip leading digits, ensure non-empty
- [x] 1.4 `SanitizeStructName`: strip all non `[a-zA-Z0-9_]` chars, title-case first char, ensure non-empty
- [x] 1.5 Add `--auto-sanitize` boolean flag to `cmd/lens-gen/main.go`
- [x] 1.6 In `main.go`, if `--auto-sanitize` is false, validate `--pkg` and `--struct` and exit 1 if invalid
- [x] 1.7 In `main.go`, if `--auto-sanitize` is true, sanitize names and add `Info` log if modified
- [x] 1.8 Test validation failure with `--pkg user-profile --struct User-profileLens`
- [x] 1.9 Test auto-sanitization with `--pkg user-profile --struct User-profileLens --auto-sanitize`
- [x] 1.10 Test against `testdata/schemas/job-posting.schema.json` with `--pkg job-posting --auto-sanitize`
- [x] 1.11 Test against `testdata/schemas/geographical-location.schema.json` with `--pkg geographical-location --auto-sanitize`

## 2. External $ref Loader

- [x] 2.1 Create `internal/gen/loader.go` with a custom `jsonschema.LoadFunc`
- [x] 2.2 Implement URL-to-filepath mapping: extract filename from `https://example.com/<filename>` and resolve relative to root schema directory
- [x] 2.3 Register the loader in `gen.NewGenerator` via `compiler.UseLoader(...)` before `compiler.Compile`
- [x] 2.4 Test against `testdata/schemas/blog-post.schema.json` (refs `user-profile.schema.json`) and verify generation succeeds
- [x] 2.5 Test against `testdata/schemas/health-record.schema.json` (refs `user-profile.schema.json`) and verify generation succeeds
- [x] 2.6 Test against `testdata/schemas/calendar.schema.json` (refs `geographical-location.schema.json`) and verify generation succeeds
- [x] 2.7 Add error handling for missing referenced files with descriptive message including attempted path

## 3. Internal $ref Traversal

- [x] 3.1 Modify `internal/gen/walker.go` `WalkSchema` to check `sch.Ref != nil` at entry and recurse into `sch.Ref`
- [x] 3.2 Guard against self-reference (`sch.Ref == sch`) to prevent infinite recursion
- [x] 3.3 Ensure `$defs` and `definitions` containers are never walked as root paths (only via `$ref`)
- [x] 3.4 Test against `testdata/schemas/ecommerce.schema.json#OrderSchema` (uses `$defs` + `$ref` to `#ProductSchema`) and verify methods for `items.<INDEX>.name` and `items.<INDEX>.price` are generated
- [x] 3.5 Test against `testdata/schemas/device.schema.json#/definitions/smartphone` (uses `definitions` + `oneOf` + `$ref`) and verify methods for `brand`, `model`, `screenSize` are generated from the smartphone definition
- [x] 3.6 Verify no spurious methods like `GetDefinitions()` or `GetDefs()` are emitted

## 4. Integration & Regression Testing

- [x] 4.1 Create a test script or Go test that iterates all `testdata/schemas/*.json` and asserts generation succeeds for each
- [x] 4.2 Verify generated `.go` files for all 10 schemas compile with `gofmt` and `go build`
- [x] 4.3 Verify no methods are generated for array-only schemas (arrays remain unsupported per Phase 2)
- [x] 4.4 Update `docs/lens-gen.md` Phase 1 section to mark completed items as done
- [x] 4.5 Run `go test ./...` and ensure all new tests pass
