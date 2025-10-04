[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/verax)](https://goreportcard.com/report/github.com/ctx42/verax)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/ctx42/verax)
![Tests](https://github.com/ctx42/verax/actions/workflows/go.yml/badge.svg?branch=master)

# verax: Validation for Go

<!-- TOC -->
* [verax: Validation for Go](#verax-validation-for-go)
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
      * [Implementing the `verax.Rule` Interface](#implementing-the-veraxrule-interface)
      * [Reusing Existing Rules](#reusing-existing-rules)
      * [Custom Validation Functions](#custom-validation-functions)
    * [Custom Errors and Error Codes](#custom-errors-and-error-codes)
    * [Conditional Rules](#conditional-rules)
    * [Skipping Rules](#skipping-rules)
  * [List of Built-In Rules](#list-of-built-in-rules)
  * [Disclaimer](#disclaimer)
<!-- TOC -->

`verax` from Latin truthful, is a flexible and intuitive Go module for
validating data structures, including primitive types, structs, slices, arrays,
and maps. It offers a simple API to define validation rules and produce clear,
human-readable error messages as well as JSON serializable error for easy API
integration. Whether validating user input, configuration data, or complex
nested structures, `verax` simplifies enforcing constraints and ensuring data
integrity.

## Features

- **Simple API**: Validate data with `verax.Validate` or `verax.ValidateStruct` for structs.
- **Built-In Rules**: Includes multiple built-in rules like `Required`, `Min`, `Max`, `Length`, and more for common validation needs.
- **Informative Errors**: Provides human-readable errors with error codes.
- **JSON Serializable Errors**: Errors serialize to JSON for API integration.
- **Struct Tag Support**: Customize error message field names using struct tags (e.g., `json` or custom tags).
- **Validator Interface**: Implement `verax.Validator` for custom struct validation logic.
- **Complex Types**: Validate slices, arrays, and maps with aggregated error reporting.
- **Extensibility**: Create custom rules implementing `verax.Rule` interface, or by using `verax.Set` and `verax.By`.
- **Conditional Validation**: Use `verax.When` and `verax.Skip` for validation logic.

## Installation

To use `verax` in your Go project, install it with:

```bash
go get github.com/ctx42/verax
```

## Usage

### Validating Primitive Types

The `verax.Validate` function validates primitive types like `int`, `string`, 
or `float64`. Pass the value and a list of rules. Rules are evaluated in order, 
and the function returns an error for the first rule that fails.

```go
err := verax.Validate(
    45,
    verax.Required,
    verax.Min(42),
    verax.Max(44),
)

PrintError(err) // Helper for formatting error output, see (examples_test.go).
PrintJSON(err)  // Helper for formatting JSON output, see (examples_test.go).
// Output:
// ERROR:
//
// - must be no greater than 44
//
// JSON:
// {
//     "code": "ECInvThreshold",
//     "error": "must be no greater than 44"
// }
```

In this example, the value `45` is checked to be non zero-value (`Required`), 
at least `42` (`Min`), and no more than `44` (`Max`). It fails the `Max(44)` 
rule. The example also shows the descriptive error message and JSON output.

### Validating Structs

The `verax.ValidateStruct` function validates struct fields. Pass a pointer to
the struct and a list of `verax.FieldRules`, each specifying a field and its
rules. Field names in errors default to the struct field name or `json` tag, 
but can be customized with `.Tag()`.

Define a struct:

```go
type Planet struct {
	Position int    `json:"position"`
	Name     string `json:"name" solar:"planet_name"`
	Life     float64
}
```

Validate an instance:

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
// - Life: must be no less than 0
// - name: the length must be between 4 and 7
// - position: must be no greater than 8
//
// JSON:
// {
//     "Life": {
//         "code": "ECInvThreshold",
//         "error": "must be no less than 0"
//     },
//     "name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "position": {
//         "code": "ECInvThreshold",
//         "error": "must be no greater than 8"
//     }
// }
```

This validates a `Planet` struct where:
 - `Position` must be between 1 and 8, 
 - `Name` must have between 4 and 7 characters long, 
 - `Life` must be between 0.0 and 1.0. 

In the above example all fields fail, and errors are presented with names 
defined in `json` tag or struct field name if it was not defined.

#### Customizing Struct Tags

By default, error messages use struct field name for the fields in the error
messages unless there is the `json` tag defined. Use `.Tag()` to specify a 
custom tag for a field.

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

Here, the `Name` field’s error uses the `solar` tag name (`planet_name`) 
instead of the `json` tag name (`name`).

#### Implementing the Validator Interface

Structs can implement the `verax.Validator` interface to define custom 
validation logic, reusable across the application:

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

Validate the struct:

```go
planet := &Planet{9, "Mer", 0.0}

err := planet.Validate()

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - planet_name: the length must be between 4 and 7
// - position: must be no greater than 8
//
// JSON:
// {
//     "planet_name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 4 and 7"
//     },
//     "position": {
//         "code": "ECInvThreshold",
//         "error": "must be no greater than 8"
//     }
// }
```

This approach encapsulates validation logic within the struct, ideal for 
consistent validation across multiple uses.

### Validating Slices and Arrays

The `verax.Validate` supports slices and arrays of structs implementing
`verax.Validator`. Each element is validated, with errors prefixed by the index 
or key.

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
// - 2.position: must be no greater than 8
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
//         "code": "ECInvThreshold",
//         "error": "must be no greater than 8"
//     }
// }
```

Each `Planet` in the slice is validated using its `Validate` method. Errors 
include the index (e.g., `0.planet_name`).

### Validating Maps

Maps with structs implementing `verax.Validator` can be validated with
`verax.Validate`. Errors are prefixed with the map key.

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
// - x.position: must be no greater than 8
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
//         "code": "ECInvThreshold",
//         "error": "must be no greater than 8"
//     }
// }
```

Each `Planet` is validated, with errors prefixed by the map key (e.g.,
`mer.planet_name`).

### Validating Map Keys and Values

Use `verax.Map` to individually assign validators to map keys-values.  

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
    verax.Key("time", verax.Min(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))),
)

err := verax.Validate(data, MyRule)
// or
err = MyRule.Validate(data)

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - bool: must be equal to 'true'
// - float: must be no less than 4.2
// - int: must be no greater than 42
// - time: must be no less than 2025-01-01T00:00:00Z
//
// JSON:
// {
//     "bool": {
//         "code": "ECNotEqual",
//         "error": "must be equal to 'true'"
//     },
//     "float": {
//         "code": "ECInvThreshold",
//         "error": "must be no less than 4.2"
//     },
//     "int": {
//         "code": "ECInvThreshold",
//         "error": "must be no greater than 42"
//     },
//     "time": {
//         "code": "ECInvThreshold",
//         "error": "must be no less than 2025-01-01T00:00:00Z"
//     }
// }
```

This validates a map with mixed types, with errors prefixed by the key.

### Custom Rules

`verax` offers three ways to create custom validation rules:

1. Implement the `verax.Rule` Interface.
2. Reuse Existing Rules.
3. Custom Validation Functions.

#### Implementing the `verax.Rule` Interface

Create a custom rule by implementing `verax.Rule`:

```go
type UserDoesNotExistRule struct{}

func (u UserDoesNotExistRule) Validate(v any) error {
	username, err := verax.EnsureString(v)
	if err != nil {
		return verax.ErrInvType
	}

    // Check if the username exists in a database.
    
	err := fmt.Errorf("user %s already exists", username)
	return xrr.Wrap(err, xrr.WithCode("ECMustNotExist"))
}
```

Use the rule:

```go
err := verax.Validate("thor", verax.Required, UserDoesNotExistRule{})

PrintError(err)
PrintJSON(err)
// Output:
// ERROR:
//
// - user thor already existexist
//
// JSON:
// {
//     "code": "ECMustNotExist",
//     "error": "user thor already existexist"
// }
```

This is ideal for rules requiring external checks, like database queries.

#### Reusing Existing Rules

Use `verax.Set` to group rules for reuse:

```go
NameRule := verax.Set{
    verax.Required,
    verax.Length(4, 5),
}

err := NameRule.Validate("abc")
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

This creates a reusable `NameRule` for non-empty strings with length between 
four and five characters.

#### Custom Validation Functions

Define a function with signature matching `func(v any) error` and use 
`verax.By`:

```go
fn := func(v any) error {
    str, err := verax.EnsureString(v)
    if err != nil {
        return verax.ErrInvType
    }
    if str != "" && str != "abc" {
        return xrr.New("i need abc", "ECMustABC")
    }
    return nil
}

AbcRule := verax.By(fn)

err := AbcRule.Validate("xyz")
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

This is useful for simple, one-off validation logic. Notice that the function
should not return errors for nils or zero-values. This is because `verax` 
has a special rule for that: `verax.Required`.

### Custom Errors and Error Codes

Customize error messages and codes with `.Error()` and `.Code()`:

```go
custom := xrr.New("must be my favorite number", "EC42")
rule := verax.Equal(42).Error(custom)

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

Customize only the error code:

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

These methods allow tailoring errors to your application’s needs.

### Conditional Rules

Use `verax.When` for conditional rules, with an optional `Else` clause:

```go
r := Range{Start: 44, End: 42}

ErrRange := xrr.New("the end must be greater than the start", "ECRange")

err := verax.ValidateStruct(
    &r,
    verax.Field(&r.End, verax.When(r.End < r.Start, verax.Error(ErrRange))),
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

This checks if `Range.End` is less than `range.Start`, applying the error if 
true.

Most of the rules in the `verax` package have `When` method which accepts a 
condition to run it or not.

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
// - End: must be no less than 100
//
// JSON:
// {
//     "End": {
//         "code": "ECInvThreshold",
//         "error": "must be no less than 100"
//     }
// }
```

This checks if `Range.Start` is greater than 50, and if so, applies the `Min` 
rule to `Range.End`.

### Skipping Rules

Use `verax.Skip` to skip subsequent rules if a condition is met:

```go
r := Range{Start: 0, End: 0}

ErrRequiredBoth := xrr.New("both values must be set", "ECRange")

err := verax.ValidateStruct(
    &r,
    verax.Field(
        &r.End,
        verax.Skip.When(r.Start > 0 && r.End > 0),
        verax.Error(ErrRequiredBoth),
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

The error triggers only if `Range.Start` and `Range.End` are both zero.

## List of Built-In Rules

`verax` provides rules for common validation scenarios:

- `Nil`: Ensures a value is `nil`.
- `Empty`: Ensures a value is not `nil` but holds `zero-value`.
- `NotNil`: Ensures a value is not `nil`.
- `Required`: Ensures a value is not `nil` and not `zero-value`.
- `NotEmpty`: Ensures a value is not `zero-value` when `non-nil`. Allows `nil` values.
- `By`: Creates a rule from a `func(v any) error`.
- `Contain`: Checks if a value is in a list using `Equal`.
- `Each`: Applies rules to each element of an array, slice, or map.
- `Equal`: Ensures a value equals a specified value.
- `NotEqual`: Ensures a value does not equal a specified value.
- `EqualBy`: Checks equality with a custom function.
- `Error`: Always fails with a specified error.
- `In`: Ensures a value is in a specified list.
- `NotIn`: Ensures a value is not in a specified list.
- `Length`: Ensures a value’s length is within a range.
- `Map`: Validates map keys with provided rules.
- `Match`: Ensures a value matches a regular expression.
- `Min`: Ensures a value is at least a specified value.
- `Max`: Ensures a value is at most a specified value.
- `Noop`: A rule that always passes.
- `Skip`: Skips subsequent rules if a condition is met.
- `When`: Applies rules conditionally, with optional `Else`.

See [GoDoc](https://pkg.go.dev/github.com/ctx42/verax) for details.

## Disclaimer

The `verax` API was inspired by
[github.com/go-ozzo/ozzo-validation](https://github.com/go-ozzo/ozzo-validation)
but was built from scratch with a different API and features, using
[github.com/ctx42/xrr](https://github.com/ctx42/xrr) for error handling.