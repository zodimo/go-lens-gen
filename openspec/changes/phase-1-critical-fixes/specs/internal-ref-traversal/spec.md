## ADDED Requirements

### Requirement: Internal $ref pointers produce lens methods
The system SHALL follow internal `$ref` pointers and generate getters/setters for the referenced schema's properties.

#### Scenario: Reference to $defs entry
- **WHEN** `ecommerce.schema.json` contains `"$ref": "#ProductSchema"` inside `order.items`
- **AND** `ProductSchema` is defined in `$defs`
- **THEN** `lens-gen --schema ecommerce.schema.json` succeeds
- **AND** the generated lens includes methods for `order.items.0.name`, `order.items.0.price`

#### Scenario: Reference to definitions entry
- **WHEN** `device.schema.json` contains `"$ref": "#/definitions/smartphone"`
- **AND** `smartphone` is defined in `definitions`
- **THEN** `lens-gen --schema device.schema.json` succeeds
- **AND** the generated lens includes methods for `brand`, `model`, `screenSize`

#### Scenario: Self-referencing root schema
- **WHEN** a schema contains a `$ref` pointing to itself (`#`)
- **THEN** the walker SHALL NOT infinitely recurse
- **AND** generation completes successfully

### Requirement: $defs and definitions are not walked as roots
The system SHALL NOT treat `$defs` or `definitions` as instance data paths.

#### Scenario: $defs container skipped
- **WHEN** `ecommerce.schema.json` defines `$defs` at the root
- **THEN** the walker does NOT generate methods like `GetDefsName()`
- **AND** only paths reachable via `$ref` or `properties` are emitted

### Requirement: Ref resolution handles all JSON Schema reference forms
The system SHALL support `$ref` values in the forms: `#`, `#/<key>`, `#/$defs/<name>`, `#/definitions/<name>`, and `#<anchor>`.

#### Scenario: Anchor reference
- **WHEN** a schema contains `"$ref": "#ProductSchema"`
- **THEN** the walker resolves the anchor and recurses into the anchored schema
