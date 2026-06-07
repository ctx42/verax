# CLAUDE.md

This file provides guidance to AI agents (and human contributors) when working
with code in this repository.

## Project Overview

`verax` (Latin for "truthful") is a Go validation library for primitive types,
structs, slices, arrays, and maps. It produces human-readable and
JSON-serializable errors, and supports custom rules and conditional validation.

Module: `github.com/ctx42/verax`

## Commands

```bash
# Run all tests
go test ./...

# Run tests with race detection (matches CI)
go test -v -race ./...

# Run a single test
go test -run Test_FunctionName ./pkg/verax/...

# Run linter (config at tmp/.golangci.yml)
golangci-lint run --config tmp/.golangci.yml ./...
```

## Architecture

### Package Layout

- **`pkg/verax`** — Core package. Contains the `Validate` and `ValidateStruct`
  entry points, all built-in rules, the error types, and struct-tag handling.
- **`pkg/verax/rule`** — Specialized rules not suitable for the core package:
  Base64, network types (IP, email, URL, CIDR, domain), and SemVer.
- **`pkg/spec`** — Serialization layer. Provides `Spec`, `Source`, `Registry`,
  and `Builder` types so rules can be represented as data and reconstructed at
  runtime (useful for storing validation config in databases or APIs).
- **`pkg/abc`** — Empty package used only as a test fixture.
- **`internal/test`** — Shared test helpers.

### Core Interfaces

```go
// Rule is the fundamental unit of validation.
type Rule interface { Validate(have any) error }

// Validator allows a struct to validate itself.
type Validator interface { Validate() error }

// Specable allows a rule to be serialized into a Spec.
type Specable interface { Spec() spec.Spec }
```

### Error Handling

All errors use `github.com/ctx42/xrr` for structured, JSON-serializable errors
with error codes (`EC` prefix) and error domains (`ED` prefix). Validation
errors aggregate field-level errors and can be marshaled directly into API
responses.

### Rule Execution Model

- `verax.Validate(value, rules...)` stops at the first failing rule.
- `verax.ValidateStruct(s)` validates all fields and collects all errors.
- `verax.When(cond, rules...)` and `verax.Skip(cond, rules...)` provide
  conditional execution.
- `verax.Each` and `verax.Map` recurse into slices/arrays and maps.
- `verax.Set` groups rules; `verax.By` and `verax.Is` wrap plain functions.

### spec Package

`Spec` is a language-agnostic rule description (name + typed args).
`Source` represents a language-specific instantiation hint. `Registry`
maps spec names to `Builder` functions that construct `Rule` values from
a `Spec`. Rules opt in to serialization by implementing `Specable`.

Not all rules can implement `Specable`. Rules that depend on Go-specific
constructs with no cross-language equivalent (e.g. `TypeRule`, which holds
a `reflect.Type`) are intentionally excluded — there is no portable way to
represent a Go type across language boundaries.

## Key Dependencies

| Module                      | Role                                              |
|-----------------------------|---------------------------------------------------|
| `github.com/ctx42/xrr`      | Structured errors with codes and fields           |
| `github.com/ctx42/mirror`   | Reflection helpers for struct field introspection |
| `github.com/ctx42/convert`  | Type conversion utilities used in rules           |
| `github.com/ctx42/jsontype` | JSON-aware type helpers                           |
| `github.com/ctx42/testing`  | Test assertions (`assert` package)                |

### Error Return Invariant

Code **inside** this package (the built-in rules, `Validate`, `ValidateStruct`,
`Field`, `When`, `Each`, `Map`, etc.) must only ever return one of the three
verax domain error types (`*Error`, `*InternalError`, or `*FieldErrors`) or nil.
Never return a raw `errors.New`, `fmt.Errorf`, or other non-domain error from
public API functions.

User-provided extension points (`Validator`, `WithValidator`, `RuleFunc`,
`EqualFunc`, `CompareFunc`, and custom rules) are currently the **caller's
responsibility**. Callers should return proper verax domain errors from these
hooks. For now the library does **not** automatically wrap foreign errors
returned from user callbacks.

This distinction exists because many callers want to return their own error
types (or errors from other libraries) from custom validation logic. The
strong invariant above applies to the library's own implementation and the
errors that ultimately come out of `Validate*` when only built-in rules are
used.

## Code Style Notes

Follow the global CLAUDE.md conventions. Additionally for this repo:

- Rule types are typically defined as structs with a `Validate(any) error`
  method and a constructor function (e.g., `func Min(n any) *Range`).
- Sentinel errors live at package level with `Err` prefix; their codes use
  `EC` prefix constants defined in the same file.
- `Specable` implementations return a `spec.Spec` describing how to
  reconstruct the rule; see existing rules for the pattern.
- `forbidigo` bans `print` and `fmt.Print*` everywhere except
  `examples_test.go`.
