# verax

**Truthful validation for Go** — Collect all errors in one pass. Beautiful, 
JSON-serializable errors. Perfect for APIs. Powered by xrr.

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/doc/devel/release)
[![License](https://img.shields.io/github/license/ctx42/verax)](LICENSE.md)
[![pkg.go.dev](https://img.shields.io/badge/pkg.go.dev-reference-blue?logo=go)](https://pkg.go.dev/github.com/ctx42/verax)
[![Changelog](https://img.shields.io/badge/changelog-v0.6.0-green)](CHANGELOG.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/verax)](https://goreportcard.com/report/github.com/ctx42/verax)

---

### Why verax?

Go validation libraries often force you to choose between:

- Stopping at the **first** error (bad UX for users)
- Returning ugly, hard-to-parse errors

**verax** solves both: it validates **everything** in a single pass, collects *
*all** failures, and returns clean, structured errors that are ready to
serialize to JSON and send straight to your API clients.

Built on top of [`xrr`](https://github.com/ctx42/xrr) — so you get stable error
codes, rich metadata, and full compatibility with `errors.Is/As`, slog, and
more.

---

### Features

- **All-errors-in-one-pass** — `ValidateStruct` reports every field failure
- **Production-ready JSON** — Errors implement `json.Marshaler` perfectly for API responses
- **Simple, fluent API** — `Validate`, `ValidateStruct`, `Field`, and rich rules
- **Built-in rules galore** — `Required`, `Min`/`Max`, `Length`, `Match`, `In`, `Each`, `Map`, network rules, semver, and more
- **Struct tags** — Uses `json` tag by default (customizable)
- **Self-validating structs** — Implement `verax.Validator`
- **Complex data** — First-class support for slices, arrays, and maps (with indexed/key errors)
- **Fully extensible** — Custom rules via `Rule` interface, reusable `Set`s, or `By` func
- **Conditional logic** — `When`, `Skip`, and per-rule conditions
- **Error classification** — `IsValidationError`, `IsInternalError`, `IsVeraxError`
- **Serializable rules** — Most built-in rules implement `SpecableRule`; encode to JSON and rebuild at runtime via `spec.Registry`
- **xrr-powered** — Stable codes + structured metadata out of the box

---

### Quick Start

```bash
go get github.com/ctx42/verax
```

<!-- gmdoceg:pkg/verax/ExampleValidator_quick_start -->
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
// - email: must be in a valid format value
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
//         "error": "must be in a valid format value"
//     },
//     "name": {
//         "code": "ECInvLength",
//         "error": "the length must be between 2 and 50"
//     }
// }
```

---

### Table of Contents

- [Installation](#installation)
- [Validating Primitive Values](#validating-primitive-values)
- [Struct Validation](#struct-validation)
- [Slices, Arrays & Maps](#slices-arrays--maps)
- [Custom Rules](#custom-rules)
- [Conditional & Skipping Rules](#conditional--skipping-rules)
- [Working with Errors](#working-with-errors)
- [Rule Serialization](#rule-serialization)
- [Comparison with Other Libraries](#comparison-with-other-libraries)
- [Contributing](#contributing)
- [License](#license)

---

### Installation

```bash
go get github.com/ctx42/verax
```

---

### Validating Primitive Values

Use `verax.Validate` for single values. Stops at first failure.

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

---

### Struct Validation

`verax.ValidateStruct` collects **all** errors from every field.

```go
err := verax.ValidateStruct(&myStruct,
	verax.Field(&myStruct.Field1, rules...),
	verax.Field(&myStruct.Field2, rules...),
)
```

Supports custom struct tags, `Validator` interface, etc.

---

### Slices, Arrays & Maps

Full support with per-element and per-key error reporting using `Each`, `Map`,
`Key`, etc.

---

### Custom Rules

- Implement `verax.Rule`
- Reuse with `verax.Set`
- Quick functions with `verax.By`

---

### Conditional & Skipping Rules

Fine-grained control with `verax.When`, `verax.Skip`, and method chaining.

---

### Working with Errors

All errors are xrr-powered:
- Stable codes (`ECInvRange`, etc.)
- JSON marshaling
- Classification helpers (`IsValidationError`, etc.)
- Easy inspection

Perfect for API responses and logging.

---

### Rule Serialization

Most built-in rules implement `verax.SpecableRule` — they can be serialized to
JSON and reconstructed at runtime. This lets you store validation configuration
in a database or config file and rebuild rules on the fly.

**Encoding a rule to JSON:**

<!-- gmdoceg:pkg/verax/ExampleBuilders_encode -->
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

**Decoding and rebuilding:**

<!-- gmdoceg:pkg/verax/ExampleBuilders_decode -->
```go
data := []byte(`{"name":"range-rule","args":{"mode":{"type":"string","value":"min"},"value":{"type":"int","value":18}}}`)

reg := spec.NewRegistry[verax.Rule]()
reg.RegisterBuilders(verax.Builders())

restored, _ := reg.DecodeAndBuild(data)

err := verax.Validate(15, restored)
fmt.Println(err)
// Output:
// must be greater or equal to 18
```

`Set` rules are serializable too — their inner rules are encoded recursively:

```go
ageRule := verax.Set{verax.Required, verax.Min(18), verax.Max(120)}
spc, _ := ageRule.Spec()
data, _ := reg.EncodeSpec(spc)
```

**By rule with a custom function:**

`By` rules wrap a function reference. Register it as a named `Source` so it
can be resolved by name during decoding:

<!-- gmdoceg:pkg/verax/ExampleBuilders_by -->
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

// Decode and rebuild — the function is resolved by name from the registry.
restored, _ := reg.DecodeAndBuild(data)

err := verax.Validate("hi", restored)
fmt.Println(err)
// Output:
// must be at least 3 characters long
```

**Serializable built-ins:** `Absent`, `By`, `Contain`, `Each`, `Equal`,
`Fail`, `In`, `Length`, `Map`, `Match`, `Noop`, `Required`, `Skip`,
`Min`/`Max` (via `Range`), and `Set`.

`TypeRule` is intentionally excluded: it holds a `reflect.Type` with no
portable cross-language representation.

---

### Comparison with Other Libraries

| Feature                    | ozzo-validation | validator.v10 | **verax**     |
|----------------------------|-----------------|---------------|---------------|
| Collects all errors        | Yes             | No            | **Yes**       |
| JSON-serializable errors   | Manual          | Limited       | **Built-in**  |
| xrr integration (codes)    | No              | No            | **Yes**       |
| Slices/maps with details   | Yes             | Yes           | **Yes**       |
| Conditional rules          | Yes             | Yes           | **Yes**       |
| Serializable rules         | No              | No            | **Yes**       |
| Simple API                 | Good            | Complex       | **Excellent** |

---

`go get github.com/ctx42/verax`

Full API docs → [pkg.go.dev/github.com/ctx42/verax](https://pkg.go.dev/github.com/ctx42/verax)  
Changelog → [CHANGELOG.md](CHANGELOG.md)

