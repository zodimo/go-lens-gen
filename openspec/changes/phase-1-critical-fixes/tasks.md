## 1. Package & Struct Sanitization

- [ ] 1.1 Create `internal/gen/sanitize.go` with `SanitizePackageName(input string) string` and `SanitizeStructName(input string) string`
- [ ] 1.2 `SanitizePackageName`: strip all non `[a-zA-Z0-9_]` chars; if leading char is not a letter, strip leading digits; ensure result is non-empty
- [ ] 1.3 `SanitizeStructName`: strip all non `[a-zA-Z0-9_]` chars; ensure first char is uppercase (title-case); ensure result is non-empty
- [ ] 1.4 Wire sanitization into `cmd/lens-gen/main.go` before passing to `gen.NewGenerator`
- [ ] 1.5 Add `Info` log when sanitization modifies user input
- [ ] 1.6 Test against `testdata/schemas/user-profile.schema.json` with `--pkg user-profile --struct User-profileLens` and verify `gofmt` passes
- [ ] 1.7 Test against `testdata/schemas/job-posting.schema.json` with `--pkg job-posting` and verify `gofmt` passes
- [ ] 1.8 Test against `testdata/schemas/geographical-location.schema.json` with `--pkg geographical-location` and verify `gofmt` passes

## 2. External $ref Loader

- [ ] 2.1 Create `internal/gen/loader.go` with a custom `jsonschema.LoadFunc`
- [ ] 2.2 Implement URL-to-filepath mapping: extract filename from `https://example.com/<filename>` and resolve relative to root schema directory
- [ ] 2.3 Register the loader in `gen.NewGenerator` via `compiler.UseLoader(...)` before `compiler.Compile`
- [ ] 2.4 Test against `testdata/schemas/blog-post.schema.json` (refs `user-profile.schema.json`) and verify generation succeeds
- [ ] 2.5 Test against `testdata/schemas/health-record.schema.json` (refs `user-profile.schema.json`) and verify generation succeeds
- [ ] 2.6 Test against `testdata/schemas/calendar.schema.json` (refs `geographical-location.schema.json`) and verify generation succeeds
- [ ] 2.7 Add error handling for missing referenced files with descriptive message including attempted path

## 3. Internal $ref Traversal

- [ ] 3.1 Modify `internal/gen/walker.go` `WalkSchema` to check `sch.Ref != nil` at entry and recurse into `sch.Ref`
- [ ] 3.2 Guard against self-reference (`sch.Ref == sch`) to prevent infinite recursion
- [ ] 3.3 Ensure `$defs` and `definitions` containers are never walked as root paths (only via `$ref`)
- [ ] 3.4 Test against `testdata/schemas/ecommerce.schema.json` (uses `$defs` + `$ref` to `#ProductSchema`) and verify methods for `order.items.0.name` and `order.items.0.price` are generated
- [ ] 3.5 Test against `testdata/schemas/device.schema.json` (uses `definitions` + `oneOf` + `$ref`) and verify methods for `brand`, `model`, `screenSize` are generated from the smartphone definition
- [ ] 3.6 Verify no spurious methods like `GetDefinitions()` or `GetDefs()` are emitted

## 4. Integration & Regression Testing

- [ ] 4.1 Create a test script or Go test that iterates all `testdata/schemas/*.json` and asserts generation succeeds for each
- [ ] 4.2 Verify generated `.go` files for all 10 schemas compile with `gofmt` and `go build`
- [ ] 4.3 Verify no methods are generated for array-only schemas (arrays remain unsupported per Phase 2)
- [ ] 4.4 Update `docs/lens-gen.md` Phase 1 section to mark completed items as done
- [ ] 4.5 Run `go test ./...` and ensure all new tests pass
