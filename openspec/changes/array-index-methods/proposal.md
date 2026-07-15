# Proposal: Array Index Methods

## Why

The `lens-gen` walker already discovers array items from JSON Schema and emits paths like `cast.<INDEX>`, but the template engine only recognizes `<DYNAMIC_KEY_...>` placeholders. This produces broken generated code where the literal string `<INDEX>` appears in the gjson path, causing silent runtime failures.

Implementing type-safe array accessors (index-based getters, setters, length, and iteration helpers) unblocks all existing test schemas that contain arrays (`movie`, `blog-post`, `health-record`, `user-profile`, `ecommerce`) and fulfills Phase 2 of the project roadmap.

## What Changes

- **Template engine rewrite** (`internal/gen/template.go`) — add left-to-right tokenization that handles both `<DYNAMIC_KEY_...>` (string substitution) and `<INDEX>` (integer substitution) in the same path, producing correct `fmt.Sprintf` format strings and argument lists.
- **Len methods** (`internal/gen/lens_template.go`) — generate `Len{Field}()` methods using `gjson.GetBytes(l.raw, "path.#").Int()` for every array path.
- **ForEach helpers** (`internal/gen/lens_template.go`) — generate `ForEach{Field}(callback func(index int, value T) bool)` wrappers around `gjson.ForEach` for homogeneous array leaf fields.
- **Walker metadata** (`internal/gen/walker.go`) — track `ArrayBasePath` on leaf fields so ForEach helpers know which gjson array path to iterate over.
- **Documentation update** (`docs/lens-gen.md`) — mark Phase 2 Items 6 and 7 as implemented, update test schema results.

## Capabilities

### New Capabilities
- `array-index-methods`: Generate type-safe index-based accessors (`GetAt` / `SetAt`), array length (`Len`), and iteration helpers (`ForEach`) for homogeneous JSON Schema arrays.

### Modified Capabilities
<!-- No existing capability requirements are changing — this is a net-new feature. -->

## Impact

- **Generated code** — all schemas with `items` arrays will now produce working array accessors instead of silently broken code.
- **CLI** — no new flags; arrays are handled automatically during schema walking.
- **Backwards compatibility** — existing static and dynamic accessors are unchanged. Array methods are additive.
- **Test schemas** — `movie.schema.json` (cast), `blog-post.schema.json` (tags), `health-record.schema.json` (allergies, conditions, medications), `user-profile.schema.json` (interests), and `ecommerce.schema.json` (order.items) will gain new generated methods.
- **Out of scope** — tuple arrays (`prefixItems`) remain deferred to a future change pending Go 1.27 generic methods support and tooling maturity.
