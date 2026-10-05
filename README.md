[![Go](https://github.com/ctx42/verax/actions/workflows/go.yml/badge.svg)](https://github.com/ctx42/verax/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ctx42/verax.svg)](https://pkg.go.dev/github.com/ctx42/verax)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/github/license/ctx42/verax)](LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/verax)](https://goreportcard.com/report/github.com/ctx42/verax)

# verax

Go validation for values, structs, slices, and maps that reports every failure
as a structured, JSON-serializable error.

## Overview

Validation libraries often stop at the first failure or return errors that
are hard to parse. `verax` validates every struct field in one pass, collects
all failures, and returns errors with stable codes that marshal straight to
JSON for API clients.

Errors are built on [`xrr`](https://github.com/ctx42/xrr), so they carry
error codes, work with `errors.Is` and `errors.As`, and integrate with
`log/slog`.

## Features

- **All errors in one pass**: `ValidateStruct` reports every field failure.
- **JSON errors**: errors implement `json.Marshaler` and carry stable codes.
- **Built-in rules**: `Required`, `Min`/`Max`, `Length`, `Match`, `In`,
  `Equal`, `Each`, `Map`, `Contain`, and more; network, SemVer, and Base64
  rules in the `rule` package.
- **Struct tags**: field names come from the `json` tag by default; choose
  another tag per field.
- **Self-validating types**: implement `verax.Validator`.
- **Collections**: slices, arrays, and maps report errors per index or key.
- **Extensible**: custom rules through the `Rule` interface, `By` functions,
  and reusable `Set`s.
- **Conditional rules**: `When`, `Skip`, and per-rule `.When(...)`.
- **Error classification**: `IsValidationError`, `IsInternalError`, and
  `IsVeraxError`.
- **Serializable rules**: most built-in rules encode to JSON and rebuild at
  runtime through `spec.Registry`.

## Packages

| Package                      | Provides                                    |
|------------------------------|---------------------------------------------|
| [`verax`](pkg/verax)         | Validate functions, built-in rules, errors  |
| [`rule`](pkg/verax/rule)     | IP, port, DNS, domain, host, SemVer, Base64 |
| [`spec`](pkg/spec/README.md) | Rule serialization to and from JSON         |

## Prerequisites

- Go 1.26 or newer.

## Installation

```bash
go get github.com/ctx42/verax
```

```go
import (
	"github.com/ctx42/verax/pkg/spec"
	"github.com/ctx42/verax/pkg/verax"
	"github.com/ctx42/verax/pkg/verax/rule"
)
```

## Quick Start

`CreateUserRequest` is a struct whose `Validate` method calls
`verax.ValidateStruct` with `Required`, `Length(2, 50)`, `Match`, and
`Min(18)`/`Max(120)` rules; see
[`examples_test.go`](pkg/verax/examples_test.go) for its definition.

<!-- gmmce:pkg/verax/ExampleValidator_quick_start -->
```go
req := &CreateUserRequest{
	Name:  "A",
	Email: "bad-email",
	Age:   15,
}

err := req.Validate()

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - age: must be greater or equal to 18
// - email: must match a valid format
// - name: the length must be between 2 and 50
//
// JSON:
// {
//     "age": {
//         "code": "ECInvRange",
//         "error": "must be greater or equal to 18"
//     },
//     "email": {
//         "code": "ECInvMatch",
//         "error": "must match a valid format"
//     },
//     "name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 2 and 50"
//     }
// }
```

The examples print errors with two helpers, `PrintError` and `PrintJSON`,
defined at the bottom of [`examples_test.go`](pkg/verax/examples_test.go),
where the `Planet` type used below is defined too.

## Usage

### Validate a value

`verax.Validate` runs rules in order and stops at the first failure.

<!-- gmmce:pkg/verax/ExampleValidate_primitive_int -->
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

### Validate a struct

`verax.ValidateStruct` runs every field's rules and collects all failures.

<!-- gmmce:pkg/verax/ExampleValidateStruct -->
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

Use another struct tag for a field's error key:

<!-- gmmce:pkg/verax/ExampleValidateStruct_custom_tag -->
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

Types implementing `verax.Validator` validate themselves:

<!-- gmmce:pkg/verax/ExampleValidator -->
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

### Slices, arrays, and maps

Elements implementing `Validator` are validated, with errors keyed by index:

<!-- gmmce:pkg/verax/ExampleValidate_slices -->
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

`Map` and `Key` validate individual map keys:

<!-- gmmce:pkg/verax/ExampleMap -->
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

err := verax.Validate(data, MyRule) //nolint:ineffassign
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

### Custom rules

Wrap a function with `By`:

<!-- gmmce:pkg/verax/ExampleBy -->
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

err := AbcRule.Validate("xyz") //nolint:ineffassign
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

Or implement the `Rule` interface, here with `UserDoesNotExistRule` from
[`examples_test.go`](pkg/verax/examples_test.go):

<!-- gmmce:pkg/verax/ExampleRule -->
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

Group rules for reuse with `Set`:

<!-- gmmce:pkg/verax/ExampleSet -->
```go
NameRule := verax.Set{
	verax.Required,
	verax.Length(4, 5),
}

err := NameRule.Validate("abc") //nolint:ineffassign
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

### Conditional rules

Run rules only when a condition holds with `When`:

<!-- gmmce:pkg/verax/ExampleWhen -->
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

Skip the remaining rules with `Skip`:

<!-- gmmce:pkg/verax/ExampleSkip -->
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

Most rules also take their own condition:

<!-- gmmce:pkg/verax/ExampleRule_conditioned -->
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

### Custom messages and codes

<!-- gmmce:pkg/verax/ExampleRule_customMessage -->
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

### Working with errors

Validation returns one of three error types, all built on `xrr`:

| Type             | Meaning                          | Check               |
|------------------|----------------------------------|---------------------|
| `*Error`         | A value failed a rule            | `IsValidationError` |
| `*FieldErrors`   | Fields, keys, or elements failed | `IsValidationError` |
| `*InternalError` | A rule was misused               | `IsInternalError`   |

`IsVeraxError` matches all three. Every error carries a stable code, such as
`ECInvRange`, and marshals to JSON as shown in the examples above.

### Rule serialization

Most built-in rules implement `verax.SpecableRule`: they encode to JSON and
rebuild at runtime, so validation configuration can live in a database or a
config file.

Encode a rule:

<!-- gmmce:pkg/verax/ExampleBuilders_encode -->
```go
rule := verax.Min(18)

reg := spec.NewRegistry[verax.Rule]()
reg.RegisterBuilders(verax.Builders())

spc, _ := rule.Spec()
data, _ := reg.EncodeSpec(spc)

fmt.Println(string(data))
// Output:
// {"name":"range-rule","args":{"mode":"min","value":{"type":"int","value":18}}}
```

Decode and rebuild it:

<!-- gmmce:pkg/verax/ExampleBuilders_decode -->
```go
data := []byte(`{
	"name": "range-rule",
	"args": {
		"mode": {"type": "string", "value": "min"},
		"value": {"type": "int", "value": 18}
	}
}`)

reg := spec.NewRegistry[verax.Rule]()
reg.RegisterBuilders(verax.Builders())

restored, _ := reg.DecodeAndBuild(data)

err := verax.Validate(15, restored)
fmt.Println(err)
// Output:
// must be greater or equal to 18
```

`By` rules wrap a function, so register it as a named `Source` to resolve it
by name when decoding:

<!-- gmmce:pkg/verax/ExampleBuilders_by -->
```go
// Register the function as a named source so it survives serialization.
src, _ := spec.NewSource("nonEmptyWord", nonEmptyWord)

reg := spec.NewRegistry[verax.Rule]()
reg.RegisterBuilders(verax.Builders())
reg.RegisterSource(src)

// Encode.
rule := verax.By(nonEmptyWord)
spc, _ := rule.Spec()
data, _ := reg.EncodeSpec(spc)

// Decode and rebuild — the function is resolved by name from the
// registry.
restored, _ := reg.DecodeAndBuild(data)

err := verax.Validate("hi", restored)
fmt.Println(err)
// Output:
// must be at least 3 characters long
```

Serializable built-ins: `Absent`, `By`, `Contain`, `Each`, `Equal`, `Fail`,
`In`, `Length`, `Map`, `Match`, `Noop`, `Required`, `Skip`, `Min`/`Max` (via
`Range`), and `Set`, whose inner rules are encoded recursively.

`TypeRule` is excluded: it holds a `reflect.Type`, which has no portable
cross-language representation.

## Resources

- [API reference](https://pkg.go.dev/github.com/ctx42/verax)
- [`spec` package](pkg/spec/README.md)
- [`xrr`](https://github.com/ctx42/xrr), the error library behind `verax`
- [Changelog](CHANGELOG.md)

## License

MIT, see [LICENSE.md](LICENSE.md).
