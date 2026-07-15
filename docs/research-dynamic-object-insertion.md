# Research: Dynamic Object Insertion with Conditional Schema Variants

## Date: 2026-07-15
## Context: Enhancing `go-lens-gen` to support inserting structured objects into dynamic maps (`additionalProperties`) with conditional schema variants (`dependencies` + `oneOf`).

---

## Executive Summary

The `database.schema.json` test schema defines a `logical` property as an object map (`additionalProperties`) where each entry is an object with a `dbType` discriminator that controls which additional fields are valid. Currently, the generator only produces **leaf field getters/setters** for dynamic objects (e.g., `GetLogicalName(key)`, `SetLogicalName(key, value)`). There is no way to **insert an entirely new logical entry** as a complete, valid object.

This document captures research into three areas:
1. **Library integration:** How `go-maybe`, `sjson`, and `GOEXPERIMENT=jsonv2` interact.
2. **Reference patterns:** How established Go libraries model unions/variants.
3. **Design options:** Four approaches explored, with a recommendation.

---

## Part 1: The Problem Space

### Current Generator Output (Existing)

For `database.schema.json`, the walker discovers the `additionalProperties` object and generates dynamic key methods:

```go
// These already exist today
func (l *DatabaseLens) GetLogicalName(logicalKey0 string) string
func (l *DatabaseLens) SetLogicalName(logicalKey0 string, value string) error
func (l *DatabaseLens) GetLogicalDbType(logicalKey0 string) string
func (l *DatabaseLens) SetLogicalDbType(logicalKey0 string, value string) error
// ... plus conditional fields (blocksize, connection, etc.)
```

### The Gap

You can modify individual fields on an **existing** logical entry, but you cannot **insert a new entry as a complete object**. sjson's `SetBytes` creates intermediate paths automatically when setting leaf values, but it only sets the leaf — it does not ensure the object satisfies schema constraints (required fields, conditional dependencies).

If you call:

```go
lens.SetLogicalName("newdb", "MyDB")
```

sjson creates `{"logical": {"newdb": {"name": "MyDB"}}}`. But `dbType` (required) is missing, and conditional fields like `blocksize` or `connection` are uninitialized.

**Desired API:**

```go
entry := database.NewLogicalEntryMemory("MyDB", maybe.Some[int64](4096))
lens.InsertLogical("newdb", entry)
```

---

## Part 2: Library Integration Research

### 2.1 go-maybe (`github.com/zodimo/go-maybe`)

**What it is:** A zero-dependency generic `Maybe`/`Option` type for Go.

**Core type:**
```go
type Maybe[T any] struct {
    value    T
    hasValue bool
}
```

**Constructors:**
```go
func Some[T any](value T) Maybe[T]  // Present value
func None[T any]() Maybe[T]          // Absent value
```

**JSON v2 support (build-tagged):**
```go
//go:build goexperiment.jsonv2
```

The library implements:
- `jsontext.MarshalerTo` — `MarshalJSONTo(*jsontext.Encoder) error`
- `jsontext.UnmarshalerFrom` — `UnmarshalJSONFrom(*jsontext.Decoder) error`
- `IsZero() bool` — Enables `json:"...,omitzero"` behavior

**Serialization behavior:**

| Value | `MarshalJSONTo` output | `omitzero` behavior |
|---|---|---|
| `Some("Alice")` | `"Alice"` | Included |
| `None[string]()` | `nil` (nothing written) | **Omitted entirely** |
| `Some(0)` | `0` | Included |
| `Some(false)` | `false` | Included |
| `Some("")` | `""` | Included |

**Key advantage over `*T` + `omitempty`:**
- `*string` with `omitempty` cannot distinguish "absent" from "present but empty string"
- `Maybe[string]` with `omitzero` correctly keeps `Some("")` in JSON while omitting `None()`

**Critical constraint:** The JSON v2 support is **completely compiled out** without `GOEXPERIMENT=jsonv2`. Without the experiment, `Maybe[T]` has no JSON marshaling at all — `encoding/json` would see unexported fields and produce `{}`.

### 2.2 GOEXPERIMENT=jsonv2

**Status:** Experimental in Go 1.25/1.26. A working group was established Nov 2025. The formal proposal (#71497) was accepted and milestone'd for **Go 1.27** (expected stable).

**What it provides:**
- `encoding/json/v2` — new high-level API with generic `MarshalFunc[T]`/`UnmarshalFunc[T]`
- `encoding/json/jsontext` — low-level streaming tokenizer
- When enabled, the existing `encoding/json` package is backed by v2 internally

**Key improvement for our use case:** `MarshalFunc[T]` allows caller-specified JSON behavior for ANY type (including generic types like `Maybe[T]`) at the call site. This is impossible in v1.

**For our generator:** If we require `GOEXPERIMENT=jsonv2`, then `sjson.SetBytes` → `encoding/json.Marshal` → v2 engine → respects `MarshalJSONTo` on `Maybe[T]`. The serialization chain works.

### 2.3 sjson Struct Handling

**Core mechanism:** When `sjson.SetBytes` receives a non-primitive value, it falls through to the `default` case:

```go
// sjson.go
default:
    b, merr := jsongo.Marshal(value)  // jsongo = encoding/json
    raw := *(*string)(unsafe.Pointer(&b))
    res, err = set(jstr, path, raw, ...)
```

**Findings:**
1. sjson uses `encoding/json.Marshal` for all structs
2. Custom `json.Marshaler` implementations are respected
3. Error propagation works (bad `MarshalJSON` output is caught)
4. Pre-marshaling with `json.Marshal` + `sjson.SetRawBytes` produces identical results
5. Deep nesting, embedded structs, and slices of structs all work

**Implication for our design:** `sjson.SetBytes(l.raw, path, entry)` where `entry` is a struct (or interface wrapping a struct) will correctly marshal the struct. As long as the struct's fields serialize correctly (via `maybe.Maybe[T]` with v2), sjson handles the rest.

---

## Part 3: Reference Pattern Research

### 3.1 OpenAI SDK (`openai-go`) — Union Types

The OpenAI SDK (generated from OpenAPI by Stainless) handles `oneOf` in two ways:

#### Request Unions (active variant pattern)

```go
// Only one field can be non-zero.
type ChatCompletionMessageParamUnion struct {
    OfDeveloper *ChatCompletionDeveloperMessageParam `json:",omitzero,inline"`
    OfSystem    *ChatCompletionSystemMessageParam    `json:",omitzero,inline"`
    OfUser      *ChatCompletionUserMessageParam      `json:",omitzero,inline"`
    OfAssistant *ChatCompletionAssistantMessageParam   `json:",omitzero,inline"`
    // ...
    paramUnion  // metadata embed
}

func (u ChatCompletionMessageParamUnion) MarshalJSON() ([]byte, error) {
    return param.MarshalUnion(u, u.OfDeveloper, u.OfSystem, ...)
}
```

**Mechanism:**
- Single struct holds pointers to all variant types
- `param.MarshalUnion()` checks `IsOmitted()` on each variant
- Serializes only the active (non-omitted) variant
- Constructor helpers: `ChatCompletionMessageParamOfFunction(content, name)`

**Relevance to our problem:** This pattern is for "pick one of N variants" — close, but our case is "the variant is determined by a discriminator field (`dbType`) within the object itself."

#### Response Unions (flattened struct pattern)

For API responses, OpenAI uses a flattened struct with ALL possible fields, plus `.AsFooVariant()` methods. This is because responses may contain fields from any variant, and they want to access them without type-switching first.

**Relevance:** Not directly applicable — our use case is constructing valid request payloads, not parsing polymorphic responses.

### 3.2 Protobuf-Go — `oneof` Types

Protobuf generates:

```go
type TestAllTypes struct {
    xxx_hidden_OneofField isTestAllTypes_OneofField
}

type isTestAllTypes_OneofField interface {
    isTestAllTypes_OneofField()
}

type testAllTypes_OneofUint32 struct {
    OneofUint32 uint32
}
func (*testAllTypes_OneofUint32) isTestAllTypes_OneofField() {}

// Introspection methods
func (x *TestAllTypes) HasOneofUint32() bool { ... }
func (x *TestAllTypes) WhichOneofField() case_TestAllTypes_OneofField { ... }
```

**Mechanism:**
- Hidden interface field in parent struct
- Wrapper structs per variant implementing the interface
- Case type for switching (`case_TestAllTypes_OneofField`)
- Has/Clear/Which methods for runtime introspection

**Relevance:** The marker interface with unexported method (`isXxx()`) is a clean Go idiom for sealed interfaces. We can adopt this. The wrapper structs are unnecessary for JSON (protobuf needs them for wire format), but the interface pattern is valuable.

---

## Part 4: Design Options Explored

### Option A: Raw Map Insertion (Simplest)

```go
func (l *DatabaseLens) InsertLogical(logicalKey0 string, obj map[string]any) error {
    path := fmt.Sprintf("logical.%s", logicalKey0)
    updated, err := sjson.SetBytes(l.raw, path, obj)
    if err == nil { l.raw = updated }
    return err
}
```

**Pros:** Easy to generate, works with any schema shape.
**Cons:** Not type-safe. Caller can pass any map. Undermines the generator's core value proposition.

### Option B: Union Struct with `maybe.Maybe[T]` (Recommended — Type-Safe)

Generate a struct representing the `additionalProperties` schema with `maybe.Maybe[T]` for optional/conditional fields:

```go
type LogicalEntryMemory struct {
    Name      string            `json:"name"`
    DbType    string            `json:"dbType"`
    Blocksize maybe.Maybe[int64] `json:"blocksize,omitzero"`
}

type LogicalEntryPostgres struct {
    Name               string             `json:"name"`
    DbType             string             `json:"dbType"`
    Connection         maybe.Maybe[string] `json:"connection,omitzero"`
    CollationIndicator maybe.Maybe[string]   `json:"collationIndicator,omitzero"`
    EnableWaitOnLock   maybe.Maybe[bool]     `json:"enableWaitOnLock,omitzero"`
}

// Marker interface
type LogicalEntry interface {
    isLogicalEntry()
}

func (LogicalEntryMemory) isLogicalEntry() {}
func (LogicalEntryPostgres) isLogicalEntry() {}

// Insert method on lens
func (l *DatabaseLens) InsertLogical(logicalKey0 string, entry LogicalEntry) error {
    path := fmt.Sprintf("logical.%s", logicalKey0)
    updated, err := sjson.SetBytes(l.raw, path, entry)
    if err == nil { l.raw = updated }
    return err
}
```

**Pros:** Fully type-safe. Variant structs only contain valid fields for their dbType. `maybe.Maybe[T]` correctly handles optional field omission via `omitzero`.
**Cons:** Requires `GOEXPERIMENT=jsonv2` for `maybe.Maybe` JSON support. More complex walker logic to detect variants from `dependencies.oneOf`.

### Option C: Fat Struct with Constructors (Simpler Generator)

Generate a single struct with ALL possible fields, using `maybe.Maybe[T]` for everything:

```go
type LogicalEntry struct {
    Name               string             `json:"name"`
    DbType             string             `json:"dbType"`
    Blocksize          maybe.Maybe[int64] `json:"blocksize,omitzero"`
    Connection         maybe.Maybe[string] `json:"connection,omitzero"`
    CollationIndicator maybe.Maybe[string] `json:"collationIndicator,omitzero"`
    EnableWaitOnLock   maybe.Maybe[bool]   `json:"enableWaitOnLock,omitzero"`
}

func NewLogicalEntryMemory(name string, blocksize maybe.Maybe[int64]) LogicalEntry {
    return LogicalEntry{Name: name, DbType: "MEMORY", Blocksize: blocksize}
}
```

**Pros:** Much simpler generator. No walker changes needed. Single struct.
**Cons:** `LogicalEntry` has fields that shouldn't be set for MEMORY (connection, etc.). No compile-time enforcement. You could accidentally set `Connection` on a MEMORY entry.

### Option D: Builder Pattern (Most Flexible, Overkill)

Fluent API with builder methods. Significant generator complexity. Probably overkill for this tool.

---

## Part 5: Recommendation

### Primary Recommendation: Option B with Multiple Structs + Marker Interface

**Rationale:**
1. The generator's core value is **type-safe** JSON access. A raw map (Option A) or fat struct (Option C) undermines this.
2. The marker interface pattern (from protobuf-go) is a clean, idiomatic Go way to model sealed variants.
3. `maybe.Maybe[T]` with `omitzero` is the most principled way to handle optional fields in Go today — it avoids the `*T` + `omitempty` pitfalls.

### Precondition: `GOEXPERIMENT=jsonv2`

**This is a hard requirement.** The generated code will not compile or function correctly without `GOEXPERIMENT=jsonv2` because `go-maybe`'s JSON support is build-tagged behind this experiment.

**Mitigation strategies if this is unacceptable:**
- Use `*T` pointers with `omitempty` (v1-compatible, semantically weaker)
- Generate a custom `Optional[T]` type with v1 `MarshalJSON`/`UnmarshalJSON` methods (more code to maintain)
- Provide a generator flag to choose between strategies

### Scope: `additionalProperties` Objects Only

The rule for generating insert structs:

> For every path where `additionalProperties` (or `patternProperties`) defines an **object-typed schema**, generate:
> 1. A marker interface
> 2. One or more struct variants (if the schema has conditional branches like `dependencies.oneOf`)
> 3. An `InsertXxx(key, entry)` method on the lens

If `additionalProperties` defines a primitive (e.g., `{"type": "string"}`), no insert method is needed — `SetXxx(key, value)` already works and sjson creates the object automatically.

---

## Part 6: Open Questions & Risks

### 1. How does the walker discover variants?

The current walker flattens `dependencies` into the same namespace:

```go
for depKey, depValue := range sch.Dependencies {
    if depSchema, ok := depValue.(*jsonschema.Schema); ok {
        WalkSchema(depSchema, currentPath, fields)  // Flattens
    }
}
```

To generate `LogicalEntryMemory` vs `LogicalEntryPostgres`, the walker needs a **second pass** or a **parallel tree** that tracks which fields came from which `oneOf` branch.

**Options:**
- Post-process the flat `fields` map after `WalkSchema` finishes, reconstructing variants by analyzing provenance
- Enhance `WalkSchema` to collect variant groupings during traversal

The post-processing approach is preferred because it keeps the walker simpler.

### 2. Default values from schema

The schema defines defaults:
```json
"dbType": { "default": "" },
"blocksize": { "default": 0 },
"collationIndicator": { "default": "7" },
"enableWaitOnLock": { "default": true }
```

If `maybe.Maybe` omits absent fields, the consumer (system reading the JSON) must apply defaults during validation. Alternatively, the generator could pre-fill defaults in constructors:

```go
func NewLogicalEntryPostgres(name string) LogicalEntryPostgres {
    return LogicalEntryPostgres{
        Name:               name,
        CollationIndicator: maybe.Some("7"),
        EnableWaitOnLock:   maybe.Some(true),
    }
}
```

### 3. Discriminator field semantics

In our schema, `dbType` is both a regular property AND the discriminator for conditional fields. Should the generator:
- Treat `dbType` as a regular string field (user sets it manually)
- Bake it into the variant struct as a constant (e.g., `DbType string` initialized in constructor)
- Generate a union type where the variant itself implies the dbType

The current lean: include `DbType` as a field but set it in constructors. This is explicit and matches JSON schema semantics.

### 4. Generalization beyond `dependencies`

What about `oneOf` at the top level of an object (not inside `dependencies`)? What about `anyOf`? What about nested `oneOf` inside `properties`?

**Proposed scope:** Start with `dependencies.{key}.oneOf` because it's the most common pattern for conditional subtyping. Expand to other patterns as needed.

### 5. `patternProperties`

The walker has a commented-out section for `PatternProperties`. If we support `additionalProperties` object inserts, we should probably also support `patternProperties` object inserts. Defer to a future phase.

---

## Part 7: Related Patterns to Consider

### Array Append (Parallel Problem)

The same structural problem exists for arrays. The generator produces `SetCastAt(index, value)` but no `AppendCast(value)`. Array append and dynamic object insert are structurally similar — both require creating new elements in collections. Consider solving both together.

### Read Support for Dynamic Objects

Currently, the generator only supports reading **leaf fields** from dynamic objects. Should we also generate:
- `GetLogicalRaw(logicalKey0 string) []byte` — returns raw JSON for the entry
- `GetLogical(logicalKey0 string) LogicalEntry` — unmarshals into the right variant

The unmarshaling path is significantly harder because it requires JSON → struct deserialization with variant detection. The `GetXxxRaw` approach is simpler and gives users flexibility.

---

## Part 8: Next Steps

### Immediate (Research Complete)

1. ✅ Document findings (this document)
2. ✅ Confirm library compatibility (`go-maybe` + `sjson` + `GOEXPERIMENT=jsonv2`)
3. ✅ Analyze reference patterns (OpenAI SDK, protobuf-go)

### Short-Term (Design Phase)

4. ⬜ Decide on `GOEXPERIMENT=jsonv2` requirement vs. fallback strategies
5. ⬜ Design walker enhancement for variant detection
6. ⬜ Sketch template additions for struct + interface generation
7. ⬜ Define scope rules (which schemas trigger variant generation)

### Medium-Term (Implementation)

8. ⬜ Implement walker post-processing for `dependencies.oneOf`
9. ⬜ Extend template to generate variant structs + marker interfaces
10. ⬜ Generate `InsertXxx` methods on the lens
11. ⬜ Generate constructor helpers for variant structs
12. ⬜ Add e2e tests with real JSON payloads
13. ⬜ Add schema test cases for conditional objects

### Long-Term (Generalization)

14. ⬜ Extend to `patternProperties`
15. ⬜ Extend to top-level `oneOf` / `anyOf` on objects
16. ⬜ Add array append methods
17. ⬜ Add read support for dynamic objects (`GetXxxRaw`)

---

## Appendix A: Concrete Generated Code (Target)

This is what the generator should produce for `database.schema.json`:

```go
package database

import (
    "fmt"
    "github.com/tidwall/gjson"
    "github.com/tidwall/sjson"
    "github.com/zodimo/go-maybe"
)

// -------------------------------------------------------------------------
// Logical Entry Types (from additionalProperties object)
// -------------------------------------------------------------------------

// LogicalEntry is a sealed interface for valid logical entry shapes.
type LogicalEntry interface {
    isLogicalEntry()
}

// LogicalEntryMemory is valid when dbType = "MEMORY"
type LogicalEntryMemory struct {
    Name      string            `json:"name"`
    DbType    string            `json:"dbType"`
    Blocksize maybe.Maybe[int64] `json:"blocksize,omitzero"`
}

func (LogicalEntryMemory) isLogicalEntry() {}

func NewLogicalEntryMemory(name string, blocksize maybe.Maybe[int64]) LogicalEntryMemory {
    return LogicalEntryMemory{
        Name:      name,
        DbType:    "MEMORY",
        Blocksize: blocksize,
    }
}

// LogicalEntryPostgres is valid when dbType = "POSTGRES" or "ORACLE"
type LogicalEntryPostgres struct {
    Name               string             `json:"name"`
    DbType             string             `json:"dbType"`
    Connection         maybe.Maybe[string] `json:"connection,omitzero"`
    CollationIndicator maybe.Maybe[string]   `json:"collationIndicator,omitzero"`
    EnableWaitOnLock   maybe.Maybe[bool]     `json:"enableWaitOnLock,omitzero"`
}

func (LogicalEntryPostgres) isLogicalEntry() {}

func NewLogicalEntryPostgres(
    name string,
    connection maybe.Maybe[string],
    collationIndicator maybe.Maybe[string],
    enableWaitOnLock maybe.Maybe[bool],
) LogicalEntryPostgres {
    return LogicalEntryPostgres{
        Name:               name,
        DbType:             "POSTGRES",
        Connection:         connection,
        CollationIndicator: collationIndicator,
        EnableWaitOnLock:   enableWaitOnLock,
    }
}

// -------------------------------------------------------------------------
// Lens Insert Method
// -------------------------------------------------------------------------

func (l *DatabaseLens) InsertLogical(logicalKey0 string, entry LogicalEntry) error {
    path := fmt.Sprintf("logical.%s", logicalKey0)
    updated, err := sjson.SetBytes(l.raw, path, entry)
    if err == nil {
        l.raw = updated
    }
    return err
}
```

---

## Appendix B: Usage Example

```go
package main

import (
    "fmt"
    "github.com/zodimo/go-maybe"
    "myproject/database"
)

func main() {
    payload := []byte(`{"logical": {"existing": {"name": "OldDB", "dbType": "MEMORY"}}}`)
    lens := database.NewDatabaseLens(payload)

    // Insert a new MEMORY entry
    entry := database.NewLogicalEntryMemory("CacheDB", maybe.Some[int64](4096))
    if err := lens.InsertLogical("cache", entry); err != nil {
        panic(err)
    }

    // Insert a new POSTGRES entry
    pgEntry := database.NewLogicalEntryPostgres(
        "ProdDB",
        maybe.Some("postgresql://localhost:5432/prod"),
        maybe.Some("8"),
        maybe.Some(true),
    )
    if err := lens.InsertLogical("prod", pgEntry); err != nil {
        panic(err)
    }

    fmt.Println(string(lens.Bytes()))
}
```

---

*Document generated during explore mode. Not yet implemented. See Part 8 for next steps.*
