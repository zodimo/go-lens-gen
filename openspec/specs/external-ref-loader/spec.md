# External Reference Loader

## Purpose
Enable lens-gen to resolve external JSON Schema `$ref` pointers to local files.

## Requirements

### Requirement: External file references compile successfully
The system SHALL resolve `$ref` pointers that reference other local schema files and produce a compiled schema tree.

#### Scenario: Cross-file reference in same directory
- **WHEN** `blog-post.schema.json` contains `"$ref": "https://example.com/user-profile.schema.json"`
- **AND** `user-profile.schema.json` exists in the same directory
- **THEN** `lens-gen --schema blog-post.schema.json` succeeds
- **AND** the generated lens includes methods for `author.username`, `author.email`, etc.

#### Scenario: Cross-file reference in dependent schema
- **WHEN** `health-record.schema.json` contains `"$ref": "https://example.com/user-profile.schema.json"`
- **AND** `user-profile.schema.json` exists in the same directory
- **THEN** `lens-gen --schema health-record.schema.json` succeeds
- **AND** the generated lens includes methods for `emergencyContact.username`, etc.

#### Scenario: Cross-file reference in calendar schema
- **WHEN** `calendar.schema.json` contains `"$ref": "https://example.com/geographical-location.schema.json"`
- **AND** `geographical-location.schema.json` exists in the same directory
- **THEN** `lens-gen --schema calendar.schema.json` succeeds
- **AND** the generated lens includes methods for `geo.latitude`, `geo.longitude`

### Requirement: Loader resolves example.com URLs to local files
The system SHALL register a custom `jsonschema` load function that maps `https://example.com/<filename>` to `<schema-dir>/<filename>`.

#### Scenario: URL to local file mapping
- **WHEN** the loader receives `https://example.com/address.schema.json`
- **AND** the root schema is in `/project/schemas/`
- **THEN** the loader reads `/project/schemas/address.schema.json`

### Requirement: Unresolvable references fail gracefully
The system SHALL produce a clear error message when an external `$ref` cannot be resolved.

#### Scenario: Missing referenced file
- **WHEN** a schema references `https://example.com/missing.schema.json`
- **AND** no such file exists in the schema directory
- **THEN** compilation fails with a descriptive error
- **AND** the error message includes the attempted file path
