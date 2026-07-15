# Design: Array Index Methods

## Context

`lens-gen` generates Go structs that provide type-safe surgical access to JSON payloads via `gjson`/`sjson`. The current walker discovers paths from JSON Schema but only supports static properties (`properties`) and dynamic map keys (`additionalProperties`).

For arrays, the walker already emits placeholder paths like `cast.<INDEX>`, but the template engine is blind to them. The regex `dynamicKeyRegex` only matches `<DYNAMIC_KEY_...>` tokens. `<INDEX>` passes through untouched, producing invalid generated code:

```go
path := "cast.<INDEX>"  // gjson looks for a literal key named "<INDEX>"
```

This is a silent runtime bug — the code compiles but fails at execution.

## Goals / Non-Goals

**Goals:**
- Generate working `Get{Field}At(index int)` / `Set{Field}At(index int, value T)` for every homogeneous array leaf field
- Generate `Len{Field}()` for every array path using `gjson`'s `.#` syntax
- Generate `ForEach{Field}(callback func(index int, value T) bool)` for homogeneous array leaf fields
- Support nested arrays (e.g., `items.items`) with multiple `indexN` arguments
- Support arrays inside dynamic maps (e.g., `users.<DYNAMIC_KEY>.tags.<INDEX>.name`)

**Non-Goals:**
- Tuple arrays (`prefixItems`) — each index has a different type. Deferred to a future change.
- Array bounds validation in generated setters
- `ForEach` for non-leaf paths (e.g., iterating over array-of-objects as whole objects)
- Array item insertion/deletion (only get/set/len/iterate)

## Decisions

### 1. Single-pass left-to-right tokenization

**Decision:** Replace the current two-step regex approach (find-all dynamic keys, then replace-all) with a single left-to-right scan that processes both `<DYNAMIC_KEY_...>` and `<INDEX>` tokens in path order.

**Rationale:** Mixed paths like `users.<DYNAMIC_KEY_users_0>.tags.<INDEX>.name` need arguments in the correct order: `usersKey0 string` then `index0 int`. The current code produces `%s` then `%s` (both strings) because it only knows about dynamic keys. A single pass ensures format string placeholders (`%s` vs `%d`) and argument lists stay aligned.

**Alternative considered:** Two separate regex passes (dynamic keys first, then indices). Rejected because argument ordering becomes ambiguous when both appear in the same path.

### 2. `At` suffix for array methods

**Decision:** Array accessor method names end with `At` (e.g., `GetCastAt`, `SetCastAt`). The `At` suffix is appended once at the end of the PascalCase path name.

**Rationale:**
- Reads naturally: "get cast at [index]"
- Distinct from static accessors (`GetCast` would be the whole array as a `gjson.Result`, which we don't generate)
- One suffix regardless of array nesting depth — the Go signature (`index0, index1 int`) makes nesting explicit

**Alternative considered:** `GetCast0`, `GetCast1` for tuples. Rejected because this change is for homogeneous arrays where a single `At` method handles all indices.

### 3. `Len{Field}()` naming and return type

**Decision:** Array length methods are named `Len{Field}()` and return `int64` to match `gjson.Result.Int()`.

**Rationale:** `int64` is consistent with integer getters. Using `int` would require a cast from `int64` and might truncate on 32-bit architectures (though unlikely for JSON array lengths).

### 4. `ForEach` only for leaf fields inside homogeneous arrays

**Decision:** `ForEach` is generated only for primitive leaf fields (`string`, `int64`, `float64`, `bool`) that live inside a homogeneous array path.

**Rationale:**
- `gjson.ForEach` iterates over values, not structured objects
- A callback for a leaf field is unambiguous: `func(index int, value string) bool`
- For array-of-objects, users can loop with `Len` + index accessors per field

**Alternative considered:** Generate a composite `ForEachItem(callback func(itemLens *ItemLens) bool)`. Rejected — requires sub-lens generation, a larger architectural change.

### 5. No index guarding

**Decision:** Generated setters do not validate `index >= 0`. `sjson` accepts negative indices (`-1` = append to array).

**Rationale:** Delegate to `sjson` semantics. Adding guards would make the generated code opinionated and might break legitimate use cases.

### 6. Array metadata tracked at walker level

**Decision:** Add `ArrayBasePath string` to `LensField` — the path to the array itself, without the `.<INDEX>` suffix. Used by `ForEach` to call `gjson.GetBytes(l.raw, "arrayBasePath").ForEach(...)`.

**Rationale:** `ForEach` needs the array root path, not the leaf path. The walker knows this at recursion time — the array base is `currentPath` before appending `.<INDEX>`.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Mixed-token paths (`dynamic` + `index`) produce wrong argument order | Unit test with `users.<DYNAMIC_KEY>.tags.<INDEX>.name` schema |
| `ForEach` callback signature clashes with user-defined types | Keep callback inline in generated code — no exported callback type |
| Nested arrays (`items.items`) generate many `indexN` args | Acceptable — call site is `GetMatrixAt(index0, index1 int) string` |
| `items: true` (boolean, accepts anything) | Walker skips — no type to infer. Documented limitation. |
| `items: false` (boolean, rejects everything) | Walker skips — schema says no items allowed. Documented limitation. |

## Migration Plan

Not applicable — this is an additive feature. Existing generated code is unaffected. Re-generating lenses from existing schemas will produce additional methods.

## Open Questions

- Should `Len` methods use `int` instead of `int64` for better ergonomics with `for i := 0; i < int(l.LenCast()); i++`?
- Should `ForEach` pass `gjson.Result` instead of the typed value for cases where users want to check `.Exists()` before conversion?
