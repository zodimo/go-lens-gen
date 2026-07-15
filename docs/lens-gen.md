# Lens-Gen: Type-Safe JSON Surgical Updater

## Overview

**Lens-Gen** is a custom schema compiler and code generator for Go. It bridges the gap between dynamic JSON APIs and the strict, type-safe world of compiled Go.

Instead of unmarshaling massive JSON payloads into heavy Go structs, Lens-Gen generates strongly typed "Lenses" that allow developers to surgically extract and mutate deeply nested data directly within the raw byte slice using `gjson` / `sjson`.

---

## Part 1: Architectural Philosophy (The "Why")

We built this tool around five core concepts:

### 1. Surgical Precision via the "Lens Pattern"

* **The Problem:** Modifying deeply nested JSON usually requires unmarshaling the entire payload into a struct, modifying one field, and marshaling it back. This risks dropping fields that aren't mapped in the struct.
* **The Solution:** We treat the JSON as an immutable byte slice. We use Lenses (`gjson` / `sjson`) to project onto specific paths, extract values, and surgically overwrite them without touching or reifying the rest of the document.

### 2. Compile-Time Safety for Dynamic Data ("Typed JSONPath")

* **The Problem:** Standard querying tools (XPath, `jq`, raw JSONPath) are "stringly typed." A typo in a path or an upstream schema type change will cause a silent runtime panic in production.
* **The Solution:** We wrap raw JSONPath strings in generated, statically-typed Go structs. If an API provider changes a schema field from a `string` to an `int`, the generated Go return type updates, and the Go compiler flags all broken business logic during the build phase.

### 3. Schema-Driven Source of Truth

* **The Problem:** Hand-writing and maintaining hundreds of JSON paths is tedious and error-prone.
* **The Solution:** We use standard JSON Schema (via `jsonschema/v5`) as the singular source of truth. The custom tree-walker ingests the schema, resolves references, infers types, and maps out the universe of possible JSON paths.

### 4. Semantic Domain Modeling

* **The Problem:** APIs frequently use dynamic dictionary keys (like UUIDs) which break static path definitions (e.g., `users.uuid-1234.status`).
* **The Solution:** When the walker encounters dynamic map properties (`additionalProperties`), it derives semantic tokens from the parent node. The template engine translates these into strongly-typed function arguments (e.g., `GetStatus(usersKey string)`), providing full IDE autocomplete that speaks the domain language.

### 5. High-Performance Execution

* **The Problem:** Large JSON unmarshaling causes massive garbage collection (GC) spikes.
* **The Solution:** By avoiding `encoding/json` and reflection entirely at runtime, the generated lenses rely purely on fast lexical scanning of raw bytes, drastically reducing GC pressure and memory allocations.

---

## Part 2: Developer How-To Guide (The "What")

### Step 1: Obtain the API Schema

You need a valid JSON Schema defining the API response. Save this file to the project root (or a designated `schemas/` directory).

*Example: `schemas/address.schema.json`*

### Step 2: Generate the Lenses

Run the `lens-gen` CLI tool, pointing it at your schema and defining the desired output package and struct name.

```bash
go run cmd/lens-gen/main.go \
  --schema schemas/address.schema.json \
  --pkg address \
  --struct AddressLens \
  --out internal/api/address/lens.go
```

**Important:** The `--pkg` flag must be a valid Go package name (no hyphens). The `--struct` flag must be a valid Go identifier.

The generator will parse the schema, walk the tree, and output a formatted `.go` file with all available Getters and Setters.

### Step 3: Integrate into the CLI Tool

You can now use the generated Lens in your business logic. You never need to define a struct or unmarshal a map.

```go
package main

import (
	"fmt"
	"log"
	"my-cli/internal/api/address" // Your generated package
)

func main() {
	// 1. Obtain raw bytes from your HTTP GET request
	rawAPIResponse := []byte(`{"streetAddress": "742 Evergreen Terrace", "locality": "Springfield"}`)

	// 2. Wrap the raw payload in the generated Lens
	lens := address.NewAddressLens(rawAPIResponse)

	// 3. Type-safe extraction (Static Path)
	locality := lens.GetLocality()
	fmt.Printf("Locality: %s\n", locality)

	// 4. Type-safe, surgical mutation
	err := lens.SetLocality("Shelbyville")
	if err != nil {
		log.Fatalf("Failed to patch JSON: %v", err)
	}

	// 5. Send the modified bytes back via HTTP POST/PUT
	finalPayload := lens.Bytes()
	_ = finalPayload
}
```

### Step 4: Handling Upstream Schema Changes

When the API provider releases a new version of their schema:

1. Download the new `schema.json` and overwrite the old one.
2. Re-run the generator command from Step 2.
3. Run `go build ./...`.
4. The Go compiler will alert you if any of your CLI's surgical updates are now invalid due to type changes or removed fields.

---

## Part 3: Supported Schema Features

The following JSON Schema features are **currently supported**:

| Feature | Status | Notes |
|---|---|---|
| `type: object` | ✅ Supported | Generates getters/setters for all leaf properties |
| `type: string` | ✅ Supported | Maps to Go `string` |
| `type: integer` | ✅ Supported | Maps to Go `int64` |
| `type: number` | ✅ Supported | Maps to Go `float64` |
| `type: boolean` | ✅ Supported | Maps to Go `bool` |
| `properties` | ✅ Supported | Static, nested object properties |
| `$id` | ✅ Supported | Schema identification |
| `$schema` | ✅ Supported | Draft declaration (2020-12) |

### Known Limitations

| Feature | Status | Current Behavior |
|---|---|---|
| `additionalProperties` | ⚠️ Partial | Walked as dynamic maps, but argument naming may collide |
| `patternProperties` | ❌ Not supported | Treated as opaque; no methods generated |
| `type: array` | ❌ Not supported | Array items are silently skipped; no index-based methods |
| `$ref` (external URL) | ❌ Not supported | Fails at compile time with "no Loader found" |
| `$ref` (internal `#anchor`) | ❌ Not supported | Resolves but walker doesn't traverse; 0 fields |
| `oneOf` | ❌ Not supported | Prints to stdout but generates no code |
| `anyOf` | ❌ Not supported | Prints to stdout but generates no code |
| `allOf` | ❌ Not supported | Prints to stdout but generates no code |
| `definitions` / `$defs` | ❌ Not supported | Schemas inside definitions are not traversed |
| `enum` | ❌ Not supported | Treated as plain primitive; no constant generation |
| `const` | ❌ Not supported | Ignored |
| `format` | ❌ Not supported | Ignored (e.g., `email`, `date`, `date-time`) |
| `required` | ❌ Not supported | No compile-time or runtime enforcement |
| `dependentRequired` | ❌ Not supported | Ignored |
| `dependencies` | ⚠️ Partial | Walked as regular properties; semantic meaning lost |
| `minimum` / `maximum` | ❌ Not supported | No bounds checking generated |
| `exclusiveMinimum` / `exclusiveMaximum` | ❌ Not supported | No bounds checking generated |
| `multipleOf` | ❌ Not supported | No validation generated |
| `minLength` / `maxLength` | ❌ Not supported | No validation generated |
| `pattern` | ❌ Not supported | No regex validation generated |
| `uniqueItems` | ❌ Not supported | No validation generated |
| `contains` | ❌ Not supported | No validation generated |
| `minContains` / `maxContains` | ❌ Not supported | No validation generated |
| `propertyNames` | ❌ Not supported | Ignored |
| `if` / `then` / `else` | ❌ Not supported | Ignored |
| `not` | ❌ Not supported | Ignored |
| `default` | ❌ Not supported | Ignored |
| `examples` | ❌ Not supported | Ignored |
| `readOnly` / `writeOnly` | ❌ Not supported | Ignored (could suppress setter for readOnly) |
| `title` / `description` | ❌ Not supported | Ignored (could be used for GoDoc) |
| `deprecated` | ❌ Not supported | Ignored |
| Package name hyphenation | ❌ Bug | `--pkg` with hyphens causes `gofmt` to fail |
| `null` type | ❌ Not supported | Not handled; would panic in walker |

### Test Schema Coverage

The project includes test schemas from [json-schema.org/learn/json-schema-examples](https://json-schema.org/learn/json-schema-examples). Current results:

| Schema | Status | Paths Found | Issue |
|---|---|---|---|
| `address.schema.json` | ✅ Pass | 7 | — |
| `movie.schema.json` | ✅ Pass | 5 | `cast` array skipped, `genre` enum treated as string |
| `user-profile.schema.json` | ❌ Fail | — | `--pkg user-profile` hyphen causes `gofmt` error |
| `blog-post.schema.json` | ❌ Fail | — | External `$ref` to `user-profile.schema.json` |
| `calendar.schema.json` | ❌ Fail | — | External `$ref` to `geographical-location.schema.json` |
| `health-record.schema.json` | ❌ Fail | — | External `$ref` to `user-profile.schema.json` |
| `device.schema.json` | ⚠️ Partial | 1 | `oneOf` and `definitions` ignored; only `deviceType` emitted |
| `ecommerce.schema.json` | ❌ Fail | 0 | `$defs` and `$ref` to `#ProductSchema` not traversed |
| `geographical-location.schema.json` | ❌ Fail | — | `--pkg geographical-location` hyphen causes `gofmt` error |
| `job-posting.schema.json` | ❌ Fail | — | `--pkg job-posting` hyphen causes `gofmt` error |

---

## Part 4: Roadmap

### Phase 1: Critical Fixes (Blocking)

1. **Package name sanitization** — Strip/replace hyphens and invalid characters in `--pkg` so `gofmt` never fails.
2. **Struct name sanitization** — Ensure `--struct` produces a valid Go identifier.
3. **External `$ref` loader** — Register a custom `jsonschema` loader that resolves relative and local file paths, or pre-process schemas into self-contained bundles.
4. **Internal `$ref` / `$defs` traversal** — When the walker encounters `$ref`, it must follow the reference and recurse into the target schema.
5. **`oneOf` / `anyOf` / `allOf` traversal** — These are currently no-ops. The walker should union the properties of all sub-schemas and generate methods for the intersection.

### Phase 2: Array Support

6. **Array index methods** — For `items`, generate `GetFieldAt(index int)` and `SetFieldAt(index int, value T)` using `gjson` array syntax (e.g., `cast.0`, `cast.1`).
7. **Array iteration helpers** — Optionally generate `ForEachField(callback)` wrappers around `gjson.ForEach`.
8. **Tuple arrays** — Support `prefixItems` (Draft 2020-12) where each index has a different type.

### Phase 3: Advanced Types & Validation

9. **`enum` support** — Generate typed constants and custom string/int types with `IsValid()` methods.
10. **Nullable types** — For `["null", "string"]`, generate `GetField() (string, bool)` or `*string` return types with `HasField()` checks.
11. **`format` code generation** — Generate runtime validators for `email`, `date`, `date-time`, `uri`, `uuid`, etc.
12. **Numeric bounds** — Generate `min`/`max` clamping or validation in setters for `minimum`, `maximum`, `exclusiveMinimum`, `multipleOf`.
13. **String constraints** — Generate `minLength`/`maxLength`/`pattern` validation in setters.

### Phase 4: Structural & Metadata

14. **`patternProperties` support** — True regex-key dynamic maps, distinct from `additionalProperties`.
15. **`dependencies` / `dependentRequired` semantics** — Understand conditional schemas and emit appropriate types.
16. **`if/then/else` support** — Generate union types or conditional accessors.
17. **`readOnly` / `writeOnly` guards** — Suppress setters for `readOnly`, suppress getters for `writeOnly`.
18. **GoDoc from `title`/`description`** — Copy schema metadata into generated method comments.
19. **Deprecation warnings** — Mark `deprecated` fields with `// Deprecated` comments.

### Phase 5: Tooling & Ergonomics

20. **Built-in `go generate` support** — Allow `//go:generate go run ...` directives.
21. **Schema bundling / dereferencing** — CLI flag to inline all `$ref` before generation.
22. **Configuration file** — `.lens-gen.yaml` for project-level defaults (package prefix, output dir, etc.).
23. **Watch mode** — Re-generate on schema file changes during development.
24. **JSON Schema test suite compliance** — Run against the official JSON Schema Test Suite to validate walker correctness.

---

## Part 5: Contributing

This project is intentionally minimal right now. The walker, template, and CLI are under 300 lines combined. Before adding features, we need **comprehensive test coverage** against the official JSON Schema Test Suite and the included example schemas.

If you want to contribute, start by picking a schema from `testdata/schemas/` that currently fails, write a failing test, and make it pass.

See `Part 3: Supported Schema Features` for the full checklist of what's left to implement.
