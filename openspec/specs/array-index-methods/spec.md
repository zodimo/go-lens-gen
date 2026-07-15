# Array Index Methods

## Purpose
Generate type-safe accessor methods for JSON Schema array fields, including index-based getters/setters, length queries, and iteration helpers.

## Requirements

### Requirement: Array index accessors are generated for homogeneous arrays
The system SHALL generate `Get{Field}At(index int)` and `Set{Field}At(index int, value T)` methods for every primitive leaf field discovered inside a homogeneous JSON Schema array (`items` as a single schema).

#### Scenario: Simple array of strings
- **WHEN** the schema defines `cast` as an array of strings (`"type": "array", "items": { "type": "string" }`)
- **THEN** the generated lens includes:
  - `func (l *MovieLens) GetCastAt(index0 int) string`
  - `func (l *MovieLens) SetCastAt(index0 int, value string) error`

#### Scenario: Array under a dynamic map key
- **WHEN** the schema defines `users.dynamicKey.tags` where `tags` is an array of strings
- **THEN** the generated lens includes:
  - `func (l *Lens) GetUsersTagsAt(usersKey0 string, index0 int) string`
  - `func (l *Lens) SetUsersTagsAt(usersKey0 string, index0 int, value string) error`

#### Scenario: Nested arrays
- **WHEN** the schema defines `matrix` as an array of arrays of strings
- **THEN** the generated lens includes:
  - `func (l *Lens) GetMatrixAt(index0, index1 int) string`
  - `func (l *Lens) SetMatrixAt(index0, index1 int, value string) error`

### Requirement: Array length methods are generated
The system SHALL generate `Len{Field}()` methods for every array path, using `gjson`'s `.#` syntax to return the array length as `int64`.

#### Scenario: Array length accessor
- **WHEN** the schema defines `cast` as an array
- **THEN** the generated lens includes:
  - `func (l *MovieLens) LenCast() int64`
- **AND** calling `LenCast()` on a payload with 3 cast members returns `3`

#### Scenario: Empty array length
- **WHEN** the array field is present but empty (`[]`)
- **THEN** `LenCast()` returns `0`

### Requirement: Array iteration helpers are generated for leaf fields
The system SHALL generate `ForEach{Field}(callback func(index int, value T) bool)` methods for every primitive leaf field inside a homogeneous array, wrapping `gjson.ForEach`.

#### Scenario: Iterating over an array of strings
- **WHEN** the schema defines `cast` as an array of strings
- **THEN** the generated lens includes:
  - `func (l *MovieLens) ForEachCast(callback func(index int, value string) bool)`
- **AND** the callback receives each array index (0, 1, 2, ...) and the string value at that index
- **AND** returning `false` from the callback stops iteration

#### Scenario: Array leaf field inside a dynamic map
- **WHEN** the schema defines `users.dynamicKey.interests` as an array of strings
- **THEN** the generated lens includes:
  - `func (l *Lens) ForEachUsersInterests(usersKey0 string, callback func(index int, value string) bool)`

### Requirement: Mixed dynamic and array tokens produce correct argument order
The system SHALL handle paths containing both `<DYNAMIC_KEY>` and `<INDEX>` tokens, producing `fmt.Sprintf` format strings and argument lists in correct left-to-right order.

#### Scenario: Dynamic key before array index
- **WHEN** the path is `users.<DYNAMIC_KEY_users_0>.tags.<INDEX>.name`
- **THEN** the generated method signature is:
  - `func (l *Lens) GetUsersTagsNameAt(usersKey0 string, index0 int) string`
- **AND** the internal path construction is:
  - `fmt.Sprintf("users.%s.tags.%d.name", usersKey0, index0)`

#### Scenario: Array index before dynamic key
- **WHEN** the path is `matrix.<INDEX>.<DYNAMIC_KEY_cells_0>.value`
- **THEN** the generated method signature is:
  - `func (l *Lens) GetMatrixCellsValueAt(index0 int, cellsKey0 string) string`
- **AND** the internal path construction is:
  - `fmt.Sprintf("matrix.%d.%s.value", index0, cellsKey0)`

### Requirement: No methods are generated for unsupported array types
The system SHALL skip array generation when `items` is a boolean (`true` or `false`) or when the schema represents a tuple array (`prefixItems`).

#### Scenario: Items is boolean true
- **WHEN** the schema defines `"items": true`
- **THEN** no array accessor methods are generated for that path
- **AND** the walker logs a debug message indicating the skip

#### Scenario: Tuple array (prefixItems)
- **WHEN** the schema defines `prefixItems`
- **THEN** no array accessor methods are generated
- **AND** a warning is logged: "Tuple arrays not yet supported — see roadmap"
