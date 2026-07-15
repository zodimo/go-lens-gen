# Package and Struct Sanitization

## Purpose
Ensure the generated package names and struct names are valid Go identifiers.

## Requirements

### Requirement: Package names are validated as valid Go identifiers
The system SHALL validate `--pkg` as a valid Go package name and reject invalid input with a descriptive error.

#### Scenario: Valid package name accepted
- **WHEN** the user runs `lens-gen --pkg userprofile`
- **THEN** generation proceeds successfully
- **AND** the generated file contains `package userprofile`

#### Scenario: Hyphenated package name rejected
- **WHEN** the user runs `lens-gen --pkg user-profile`
- **THEN** generation fails before code generation
- **AND** the error message indicates "invalid package name 'user-profile'"
- **AND** the error message explains that Go package names may only contain lowercase letters, digits, and underscores
- **AND** the error message suggests `userprofile` as a valid alternative

#### Scenario: Package name with uppercase letters rejected
- **WHEN** the user runs `lens-gen --pkg UserProfile`
- **THEN** generation fails before code generation
- **AND** the error message indicates package names must be lowercase

#### Scenario: Package name starting with digit rejected
- **WHEN** the user runs `lens-gen --pkg 123test`
- **THEN** generation fails before code generation
- **AND** the error message indicates package names must start with a letter

### Requirement: Struct names are validated as valid Go identifiers
The system SHALL validate `--struct` as a valid, exported Go identifier and reject invalid input with a descriptive error.

#### Scenario: Valid struct name accepted
- **WHEN** the user runs `lens-gen --struct CustomerLens`
- **THEN** generation proceeds successfully
- **AND** the generated file contains `type CustomerLens struct`

#### Scenario: Hyphenated struct name rejected
- **WHEN** the user runs `lens-gen --struct user-profile-lens`
- **THEN** generation fails before code generation
- **AND** the error message indicates "invalid struct name 'user-profile-lens'"
- **AND** the error message explains that Go identifiers may only contain letters, digits, and underscores
- **AND** the error message suggests `UserProfileLens` as a valid alternative

#### Scenario: Unexported struct name rejected
- **WHEN** the user runs `lens-gen --struct customerLens`
- **THEN** generation fails before code generation
- **AND** the error message indicates struct names must be exported (start with uppercase)

#### Scenario: Struct name with special characters rejected
- **WHEN** the user runs `lens-gen --struct my_lens!@#`
- **THEN** generation fails before code generation
- **AND** the error message indicates invalid characters

### Requirement: Auto-sanitization is available via opt-in flag
The system SHALL provide an `--auto-sanitize` flag that automatically transforms invalid names into valid equivalents.

#### Scenario: Auto-sanitization produces valid package name
- **WHEN** the user runs `lens-gen --pkg user-profile --auto-sanitize`
- **THEN** generation proceeds with `package userprofile`
- **AND** an `Info` log indicates the original and sanitized values

#### Scenario: Auto-sanitization produces valid struct name
- **WHEN** the user runs `lens-gen --struct user-profile-lens --auto-sanitize`
- **THEN** generation proceeds with `type UserProfileLens struct`
- **AND** an `Info` log indicates the original and sanitized values

#### Scenario: Auto-sanitization not applied to already-valid names
- **WHEN** the user runs `lens-gen --pkg userprofile --struct CustomerLens --auto-sanitize`
- **THEN** no sanitization log is emitted
- **AND** original names are used unchanged
