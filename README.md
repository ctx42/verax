[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/verax)](https://goreportcard.com/report/github.com/ctx42/verax)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/ctx42/verax)
![Tests](https://github.com/ctx42/verax/actions/workflows/go.yml/badge.svg?branch=master)

# verax: Validation for Go

<!-- TOC -->
* [verax: Validation for Go](#verax-validation-for-go)
  * [Quick Start](#quick-start)
  * [Features](#features)
  * [Installation](#installation)
  * [Usage](#usage)
    * [Validating Primitive Types](#validating-primitive-types)
    * [Validating Structs](#validating-structs)
      * [Customizing Struct Tags](#customizing-struct-tags)
      * [Implementing the Validator Interface](#implementing-the-validator-interface)
    * [Validating Slices and Arrays](#validating-slices-and-arrays)
    * [Validating Maps](#validating-maps)
    * [Validating Map Keys and Values](#validating-map-keys-and-values)
    * [Custom Rules](#custom-rules)
      * [Implementing the Rule Interface](#implementing-the-rule-interface)
      * [Reusing Existing Rules](#reusing-existing-rules)
      * [Custom Validation Functions](#custom-validation-functions)
    * [Custom Errors and Error Codes](#custom-errors-and-error-codes)
    * [Working with Validation Errors](#working-with-validation-errors)
      * [Error Types](#error-types)
      * [Error Classification](#error-classification)
      * [JSON Marshalling](#json-marshalling)
      * [Inspecting Errors Programmatically](#inspecting-errors-programmatically)
    * [Conditional Rules](#conditional-rules)
    * [Skipping Rules](#skipping-rules)
  * [List of Built-In Rules](#list-of-built-in-rules)
  * [Disclaimer](#disclaimer)
<!-- TOC -->

`verax` (Latin for "truthful") is a Go validation library for primitive types,
structs, slices, arrays, and maps. It produces human-readable,
JSON-serializable errors suitable for direct use in API responses.

## Quick Start

```go
import "github.com/ctx42/verax"

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func (r *CreateUserRequest) Validate() error {
    return verax.ValidateStruct(r,
        verax.Field(&r.Name,  verax.Required, verax.Length(2, 50)),
        verax.Field(&r.Email, verax.Required, verax.Match(emailRx)),
        verax.Field(&r.Age,   verax.Required, verax.Min(18), verax.Max(120)),
    )
}

req := &CreateUserRequest{Name: "A", Email: "bad", Age: 15}
if err := req.Validate(); err != nil {
    data, _ := json.Marshal(err)
    // {
    //   "age":   {"code": "ECInvRange",  "error": "must be greater or equal to 18"},
    //   "email": {"code": "ECInvMatch",  "error": "must be in a valid format value"},
    //   "name":  {"code": "ECInvLength", "error": "the length must be between 2 and 50"}
    // }
}
```

## Features

- **Simple API**: `verax.Validate` for values, `verax.ValidateStruct` for structs.
- **Collect All Errors**: `ValidateStruct` reports every failing field in one pass.
- **Built-In Rules**: `Required`, `Min`, `Max`, `Length`, `Equal`, `Match`, `In`, `Each`, `Map`, and more.
- **JSON-Ready Errors**: All error types implement `json.Marshaler` and `json.Unmarshaler` — marshal directly into API responses.
- **Error Classification**: `IsVeraxError`, `IsValidationError`, and `IsInternalError` distinguish error categories without type assertions.
- **Struct Tag Support**: Field names in errors default to the `json` tag; override with `.Tag()`.
- **Validator Interface**: Structs implement `verax.Validator` for self-contained, reusable validation logic.
- **Complex Types**: Validate slices, arrays, and maps with per-element error reporting.
- **Extensible**: Implement `verax.Rule`, compose with `verax.Set`, or wrap functions with `verax.By`.
- **Conditional Validation**: `verax.When`, `verax.Skip`, and per-rule `.When()` for fine-grained control.

## Installation

```bash
go get github.com/ctx42/verax
```

## Usage

### Validating Primitive Types

`verax.Validate` validates a single value against a list of rules. Rules run in
order and the function returns on the first failure. `PrintError` and `PrintJSON`
are helpers defined in `examples_test.go`.

<!-- gmdoceg:pkg/verax/ExampleValidate_primitive_int -->
```go
err := verax.Validate(
	45,
	verax.Required,
	verax.Min(42),
	verax.Max(44),
)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - must be less or equal to 44
//
// JSON:
// {
//     "code": "ECInvRange",
//     "error": "must be less or equal to 44"
// }
```

### Validating Structs

`verax.ValidateStruct` validates all fields in one pass, collecting every
failure before returning. Pass a pointer to the struct and `verax.Field`
descriptors, each specifying a field and its rules. Field names in errors
default to the `json` tag, or the struct field name when no tag is defined.

Define a struct:

```go
type Planet struct {
    Position int     `json:"position"`
    Name     string  `json:"name" solar:"planet_name"`
    Life     float64
}
```

Validate an instance:

<!-- gmdoceg:pkg/verax/ExampleValidateStruct -->
```go
planet := Planet{9, "PlanetXYZ", -1}

err := verax.ValidateStruct(
	&planet,
	verax.Field(&planet.Position, verax.Min(1), verax.Max(8)),
	verax.Field(&planet.Name, verax.Length(4, 7)),
	verax.Field(&planet.Life, verax.Min(0.0), verax.Max(1.0)),
)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - Life: must be greater or equal to 0
// - name: the length must be between 4 and 7
// - position: must be less or equal to 8
//
// JSON:
// {
//     "Life": {
//         "code": "ECInvRange",
//         "error": "must be greater or equal to 0"
//     },
//     "name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "position": {
//         "code": "ECInvRange",
//         "error": "must be less or equal to 8"
//     }
// }
```

#### Customizing Struct Tags

Use `.Tag()` to pick a different struct tag for the field name in errors:

<!-- gmdoceg:pkg/verax/ExampleValidateStruct_custom_tag -->
```go
planet := Planet{1, "Mer", 0.0}

err := verax.ValidateStruct(
	&planet,
	verax.Field(&planet.Name, verax.Length(4, 7)).Tag("solar"),
)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - planet_name: the length must be between 4 and 7
//
// JSON:
// {
//     "planet_name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     }
// }
```

#### Implementing the Validator Interface

Structs that implement `verax.Validator` are validated automatically when
passed to `verax.Validate`, or when they appear as elements in a slice, array,
or map:

```go
func (p *Planet) Validate() error {
    return verax.ValidateStruct(
        p,
        verax.Field(&p.Position, verax.Min(1), verax.Max(8)),
        verax.Field(&p.Name, verax.Length(4, 7)).Tag("solar"),
        verax.Field(&p.Life, verax.Min(0.0), verax.Max(1.0)),
    )
}
```

<!-- gmdoceg:pkg/verax/ExampleValidator -->
```go
planet := &Planet{9, "Mer", 0.0}

err := planet.Validate()

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - planet_name: the length must be between 4 and 7
// - position: must be less or equal to 8
//
// JSON:
// {
//     "planet_name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "position": {
//         "code": "ECInvRange",
//         "error": "must be less or equal to 8"
//     }
// }
```

### Validating Slices and Arrays

Slices and arrays of `verax.Validator` values are validated element by element.
Errors are prefixed with the element index.

<!-- gmdoceg:pkg/verax/ExampleValidate_slices -->
```go
planets := []*Planet{
	{1, "Mer", 0},
	{3, "Earth", 1.0},
	{9, "X", 0.1},
}

err := verax.Validate(planets)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - 0.planet_name: the length must be between 4 and 7
// - 2.planet_name: the length must be between 4 and 7
// - 2.position: must be less or equal to 8
//
// JSON:
// {
//     "0.planet_name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "2.planet_name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "2.position": {
//         "code": "ECInvRange",
//         "error": "must be less or equal to 8"
//     }
// }
```

### Validating Maps

Maps whose values implement `verax.Validator` are validated the same way.
Errors are prefixed with the map key.

<!-- gmdoceg:pkg/verax/ExampleValidate_maps -->
```go
planets := map[string]*Planet{
	"mer": {1, "Mer", 0},
	"ear": {3, "Earth", 1.0},
	"x":   {9, "X", 0.1},
}

err := verax.Validate(planets)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - mer.planet_name: the length must be between 4 and 7
// - x.planet_name: the length must be between 4 and 7
// - x.position: must be less or equal to 8
//
// JSON:
// {
//     "mer.planet_name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "x.planet_name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "x.position": {
//         "code": "ECInvRange",
//         "error": "must be less or equal to 8"
//     }
// }
```

### Validating Map Keys and Values

Use `verax.Map` to assign rules to specific map keys:

<!-- gmdoceg:pkg/verax/ExampleMap -->
```go
data := map[string]any{
	"bool":  false,
	"int":   44,
	"float": 0.1,
	"time":  time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
}

MyRule := verax.Map(
	verax.Key("bool", verax.Equal(true)),
	verax.Key("int", verax.Max(42)),
	verax.Key("float", verax.Min(4.2)),
	verax.Key("time", verax.Min(
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	)),
)

err := verax.Validate(data, MyRule) // nolint: ineffassign
// or
err = MyRule.Validate(data)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - bool: must be equal to 'true'
// - float: must be greater or equal to 4.2
// - int: must be less or equal to 42
// - time: must be greater or equal to 2025-01-01T00:00:00Z
//
// JSON:
// {
//     "bool": {
//         "code": "ECNotEqual",
//         "error": "must be equal to 'true'"
//     },
//     "float": {
//         "code": "ECInvRange",
//         "error": "must be greater or equal to 4.2"
//     },
//     "int": {
//         "code": "ECInvRange",
//         "error": "must be less or equal to 42"
//     },
//     "time": {
//         "code": "ECInvRange",
//         "error": "must be greater or equal to 2025-01-01T00:00:00Z"
//     }
// }
```

### Custom Rules

`verax` offers three ways to create custom validation rules:

1. Implement the `verax.Rule` interface.
2. Compose existing rules with `verax.Set`.
3. Wrap a plain function with `verax.By`.

#### Implementing the Rule Interface

Implement `verax.Rule` when you need full control, such as querying a database.
The `Validate` method must return:
- `nil` on success,
- `*verax.Error` for a user-facing validation failure, or
- `*verax.InternalError` for misuse (e.g. a caller passed the wrong type).

```go
type UserDoesNotExistRule struct{}

func (u UserDoesNotExistRule) Validate(have any) error {
    username, ok := have.(string)
    if !ok {
        return verax.NewInternalError("must be a string", verax.ECInvType)
    }
    if userExists(username) {
        return verax.NewError(
            fmt.Sprintf("user %s already exist", username),
            "ECMustNotExist",
        )
    }
    return nil
}
```

Use the rule:

<!-- gmdoceg:pkg/verax/ExampleRule -->
```go
err := verax.Validate("thor", verax.Required, UserDoesNotExistRule{})

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - user thor already exist
//
// JSON:
// {
//     "code": "ECMustNotExist",
//     "error": "user thor already exist"
// }
```

#### Reusing Existing Rules

Use `verax.Set` to compose rules for reuse:

<!-- gmdoceg:pkg/verax/ExampleSet -->
```go
NameRule := verax.Set{
	verax.Required,
	verax.Length(4, 5),
}

err := NameRule.Validate("abc") // nolint: ineffassign
// or
err = verax.Validate("abc", NameRule)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - the length must be between 4 and 5
//
// JSON:
// {
//     "code": "ECInvLength",
//     "error": "the length must be between 4 and 5"
// }
```

#### Custom Validation Functions

Wrap a `func(v any) error` with `verax.By`:

<!-- gmdoceg:pkg/verax/ExampleBy -->
```go
fn := func(v any) error {
	str, err := verax.EnsureString(v)
	if err != nil {
		return verax.ErrInvType
	}
	if str != "" && str != "abc" {
		return verax.NewError("i need abc", "ECMustABC")
	}
	return nil
}

AbcRule := verax.By(fn)

err := AbcRule.Validate("xyz") // nolint: ineffassign
// or
err = verax.Validate("xyz", AbcRule)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - i need abc
//
// JSON:
// {
//     "code": "ECMustABC",
//     "error": "i need abc"
// }
```

> **Tip**: Custom functions should return `nil` for nil or zero values. Use
> `verax.Required` as a separate rule to enforce non-zero values.

### Custom Errors and Error Codes

Customize error messages and codes with `.Message()` and `.Code()`:

<!-- gmdoceg:pkg/verax/ExampleRule_customMessage -->
```go
rule := verax.Equal(42).Message("must be my favorite number").Code("EC42")

err := verax.Validate(44, rule)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - must be my favorite number
//
// JSON:
// {
//     "code": "EC42",
//     "error": "must be my favorite number"
// }
```

Customize only the error code, keeping the default message:

<!-- gmdoceg:pkg/verax/ExampleRule_customErrorCode -->
```go
rule := verax.Equal(42).Code("EC42")

err := verax.Validate(44, rule)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - must be equal to '42'
//
// JSON:
// {
//     "code": "EC42",
//     "error": "must be equal to '42'"
// }
```

### Working with Validation Errors

#### Error Types

`verax` uses three distinct error types, all in the same error domain:

| Type                   | Meaning                                                            |
|------------------------|--------------------------------------------------------------------|
| `*verax.Error`         | Single-value validation failure (user-facing)                      |
| `*verax.FieldsError`   | Multi-field validation failure from `ValidateStruct` (user-facing) |
| `*verax.InternalError` | Library misuse — wrong type, incomplete rule setup, etc.           |

#### Error Classification

Three predicate functions classify errors without type assertions:

```go
err := someValidation()

// IsValidationError is true for *Error and *FieldsError — safe to return to
// the client.
if verax.IsValidationError(err) {
    // handle user-facing validation failure
}

// IsInternalError is true for *InternalError — indicates a programming
// mistake; log it and return a 500.
if verax.IsInternalError(err) {
    log.Println("bug: internal validation error:", err)
}

// IsVeraxError is true for any verax error (the union of the two above).
if verax.IsVeraxError(err) {
    // handle any verax error
}
```

The three functions satisfy the invariant:
`IsVeraxError(err) == IsValidationError(err) || IsInternalError(err)`

#### JSON Marshalling

All three error types implement `json.Marshaler` and `json.Unmarshaler`. Pass
a `verax` error directly to `json.Marshal` or any JSON encoder — no manual
conversion is needed.

A single validation error (`*verax.Error`):

```json
{"error": "must be no greater than 44", "code": "ECInvRange"}
```

With optional metadata attached via `xrr.Meta`:

```json
{"error": "must be no greater than 44", "code": "ECInvRange", "meta": {"field": "age"}}
```

A struct validation error (`*verax.FieldsError`) from `ValidateStruct`:

```json
{
    "name":     {"error": "the length must be between 4 and 7", "code": "ECInvLength"},
    "position": {"error": "must be less or equal to 8",          "code": "ECInvRange"}
}
```

A typical HTTP handler using verax errors directly:

```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    if err := req.Validate(); err != nil {
        if verax.IsInternalError(err) {
            http.Error(w, "internal error", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(err) // marshals to structured JSON automatically
        return
    }

    // process the valid request ...
}
```

Because errors also implement `json.Unmarshaler`, they can be decoded back from
JSON — useful for forwarding validation errors between services or for
client-side processing.

#### Inspecting Errors Programmatically

Use `errors.AsType` (Go 1.25+) to obtain a typed handle on a validation error:

```go
err := verax.ValidateStruct(&planet, ...)

// Iterate over field-level failures.
if fe, ok := errors.AsType[*verax.FieldsError](err); ok {
    for field, fieldErr := range fe.ErrorFields() {
        fmt.Println(field, fieldErr)
    }
}

// Read the error code from a single validation error.
if e, ok := errors.AsType[*verax.Error](err); ok {
    fmt.Println(e.ErrorCode())
}

// Handle an internal error (library misuse).
if _, ok := errors.AsType[*verax.InternalError](err); ok {
    log.Println("internal validation error:", err)
}
```

### Conditional Rules

Use `verax.When` for conditional rules, with an optional `Else` clause:

<!-- gmdoceg:pkg/verax/ExampleWhen -->
```go
r := Range{Start: 44, End: 42}

failRule := verax.Fail("the end must be greater than the start", "ECRange")

err := verax.ValidateStruct(
	&r,
	verax.Field(&r.End, verax.When(r.End < r.Start, failRule)),
)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - End: the end must be greater than the start
//
// JSON:
// {
//     "End": {
//         "code": "ECRange",
//         "error": "the end must be greater than the start"
//     }
// }
```

Most built-in rules also have a `.When(condition bool)` method for inline
conditioning:

<!-- gmdoceg:pkg/verax/ExampleRule_conditioned -->
```go
r := Range{Start: 51, End: 42}

err := verax.ValidateStruct(
	&r,
	verax.Field(&r.End, verax.Min(100).When(r.Start > 50)),
)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - End: must be greater or equal to 100
//
// JSON:
// {
//     "End": {
//         "code": "ECInvRange",
//         "error": "must be greater or equal to 100"
//     }
// }
```

### Skipping Rules

Use `verax.Skip` to skip subsequent rules when a condition is met:

<!-- gmdoceg:pkg/verax/ExampleSkip -->
```go
r := Range{Start: 0, End: 0}

err := verax.ValidateStruct(
	&r,
	verax.Field(
		&r.End,
		verax.Skip.When(r.Start > 0 && r.End > 0),
		verax.Fail("both values must be set", "ECRange"),
	),
)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - End: both values must be set
//
// JSON:
// {
//     "End": {
//         "code": "ECRange",
//         "error": "both values must be set"
//     }
// }
```

## List of Built-In Rules

| Rule                   | Description                                                                   |
|------------------------|-------------------------------------------------------------------------------|
| `Required`             | Value must not be nil and not a zero value.                                   |
| `Nil`                  | Value must be nil.                                                            |
| `NotNil`               | Value must not be nil.                                                        |
| `Empty`                | Value must be a non-nil zero value.                                           |
| `NotEmpty`             | Value must be non-zero when non-nil; nil is allowed.                          |
| `Equal(want)`          | Value must equal `want`.                                                      |
| `NotEqual(want)`       | Value must not equal `want`.                                                  |
| `Min(n)`               | Value must be ≥ `n`.                                                          |
| `Max(n)`               | Value must be ≤ `n`.                                                          |
| `Length(min, max)`     | Length must be in `[min, max]`.                                               |
| `RuneLength(min, max)` | Rune length must be in `[min, max]`; for non-strings, behaves like `Length`.  |
| `Match(rx)`            | Value must match the regular expression `rx`.                                 |
| `In(values...)`        | Value must be one of `values`.                                                |
| `NotIn(values...)`     | Value must not be in `values`.                                                |
| `Contain(values...)`   | Value must contain all of `values`.                                           |
| `Each(rules...)`       | Apply `rules` to every element of a slice, array, or map.                     |
| `Map(keys...)`         | Apply per-key rules to a map (see `verax.Key`).                               |
| `Type(t)`              | Value must be of reflect type `t`.                                            |
| `By(fn)`               | Wrap a `func(have any) error` as a rule.                                      |
| `Check(fn, msg, code)` | Wrap a typed `func[T any](v T) bool` as a rule with a fixed message and code. |
| `Set{rules...}`        | Compose rules into a reusable group.                                          |
| `Fail(msg, code)`      | Always fails with the given message and code.                                 |
| `Noop`                 | Always passes (useful as a placeholder).                                      |
| `When(cond, rules...)` | Apply `rules` only when `cond` is true.                                       |
| `Skip`                 | Skip subsequent rules; combine with `.When()` for conditions.                 |

Additional rules for network addresses (IP, IPv4, IPv6, port, DNS, domain,
host), Base64, and Semantic Versioning live in `pkg/verax/rule`.

See [GoDoc](https://pkg.go.dev/github.com/ctx42/verax) for the complete API
reference.

## Disclaimer

The `verax` API was inspired by
[github.com/go-ozzo/ozzo-validation](https://github.com/go-ozzo/ozzo-validation)
but was built from scratch with a different API and features, using
[github.com/ctx42/xrr](https://github.com/ctx42/xrr) for error handling.
