## Context

Lens-Gen is a Go code generator that produces type-safe JSON lens structs from JSON Schema files. It uses `github.com/santhosh-tekuri/jsonschema/v5` for schema compilation and `gjson`/`sjson` for runtime JSON access.

Current state:
- The CLI accepts `--pkg` and `--struct` flags and passes them directly into the Go template.
- The `jsonschema` compiler uses default settings, which cannot resolve external file references.
- The tree walker (`WalkSchema`) only recurses into `Properties`, `AdditionalProperties`, `Items`, and `Dependencies`. It does not follow `$ref` pointers or descend into `$defs`/`definitions`.
- All 10 test schemas from json-schema.org fail in some way: hyphens in names, external `$ref`, internal `$ref`/`$defs`, or `oneOf`.

## Goals / Non-Goals

**Goals:**
- Make the generator usable on all 10 test schemas without manual schema editing.
- Produce valid Go code for any reasonable `--pkg` and `--struct` input.
- Resolve external `$ref` to files in the same directory as the root schema.
- Resolve internal `$ref` (`#`, `#/definitions/x`, `#/$defs/x`, `#/$anchor`) and continue walking.

**Non-Goals:**
- Full JSON Schema compliance (validation, conditional schemas, `if/then/else`).
- Network-based `$ref` resolution (HTTP URLs are out of scope for now).
- Array support (Phase 2).
- `oneOf`/`anyOf`/`allOf` code generation (Phase 1 point 5 is explicitly excluded per user request).

## Decisions

### Decision 1: Validate `--pkg` and `--struct` strictly; offer auto-sanitization as opt-in
**Rationale:** Silent transformation violates the principle of least surprise. A developer running a Go code generator should know Go naming conventions, and explicit control over package/struct names prevents accidental collisions. Failing fast with a clear, actionable error is preferable to generating code with unexpected names.
**Behavior:**
- By default, `--pkg` and `--struct` are validated against Go identifier rules.
- If invalid, the CLI exits with code 1 and prints a descriptive error that includes a suggested valid alternative.
- An `--auto-sanitize` flag allows users to opt into automatic transformation when batch-generating from filenames.
**Alternative considered:** Auto-sanitize by default. Rejected because it hides naming decisions from the user and may produce colliding package names (`user-profile` and `user_profile` both become `userprofile`).

### Decision 2: Provide helper functions for validation and sanitization
**Rationale:** The validation and sanitization logic should live in `internal/gen` so it can be unit-tested and reused. The CLI layer is responsible for deciding whether to validate or sanitize based on flags.
**Implementation:**
- `ValidatePackageName(name string) error` — checks `[a-z][a-z0-9_]*`, returns descriptive error with suggestion.
- `ValidateStructName(name string) error` — checks `[A-Z][A-Za-z0-9_]*`, returns descriptive error with suggestion.
- `SanitizePackageName(name string) string` — strips invalid chars, lowercases, ensures leading letter.
- `SanitizeStructName(name string) string` — strips invalid chars, title-cases first letter.

### Decision 3: Register a custom `jsonschema.LoadFunc` that resolves `https://example.com/*.schema.json` to local files
**Rationale:** The test schemas use `https://example.com/` as a base URI. The `jsonschema` compiler's default loader only handles `file://` and embedded schemas. We need a loader that intercepts these URLs and resolves them relative to the schema file's directory.
**Implementation:** In `NewGenerator`, after creating the compiler, register a load function that:
1. Checks if the URL host is `example.com`.
2. Extracts the filename from the path (e.g., `user-profile.schema.json`).
3. Looks for that file in the same directory as the root schema.
4. Falls back to the default loader for other URLs.

### Decision 4: In `WalkSchema`, detect `$ref` and recurse into the resolved schema
**Rationale:** The `jsonschema/v5` library resolves `$ref` internally during compilation, so the compiled schema tree already contains the merged/resolved schema at each node. However, if a schema node has `$ref` set, its local properties might be empty because the referenced schema is attached elsewhere. The walker must check if `sch.Ref` is non-nil and recurse into it.
**Implementation:** At the top of `WalkSchema`, add:
```go
if sch.Ref != nil {
    WalkSchema(sch.Ref, currentPath, fields)
    return
}
```
This handles both external and internal references because the compiler has already resolved them.

### Decision 5: Handle `$defs` and `definitions` by treating them as schema containers, not walkable roots
**Rationale:** `$defs` and `definitions` are not part of the instance data path — they are reusable schema fragments. The walker should only encounter them via `$ref`. If the root schema has `$defs`, the walker should not iterate over them directly (they have no JSON path). The `Ref` resolution in Decision 4 handles this automatically.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Strict validation increases friction for newcomers | Error messages include suggested valid names, making the fix trivial; `--auto-sanitize` flag available for convenience |
| Auto-sanitization may produce colliding package names (e.g., `user-profile` and `user_profile` both become `userprofile`) | Document the behavior; default validation mode prevents this entirely |
| Custom loader assumes all `example.com` schemas are co-located | This is true for the test suite; for production use, a `--base-url` flag can be added later |
| Following `sch.Ref` may cause infinite recursion on circular references | The `jsonschema` compiler already detects and errors on circular `$ref`; we trust that invariant |
| Internal `$ref` to `#` (self-reference) might double-walk properties | Check if `sch.Ref == sch` (self-reference) and skip to avoid duplication |

## Migration Plan

No migration needed — this is a backward-compatible bugfix. Existing working schemas will continue to work. Schemas that previously failed will now succeed.

## Open Questions

- Should the custom loader support relative `file://` URIs in addition to `https://example.com/`? (Leaning toward yes — resolve relative to schema directory.)
