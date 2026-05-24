// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
	"github.com/ctx42/verax/pkg/verax"
)

var emailRx = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (r *CreateUserRequest) Validate() error {
	return verax.ValidateStruct(r,
		verax.Field(&r.Name, verax.Required, verax.Length(2, 50)),
		verax.Field(&r.Email, verax.Required, verax.Match(emailRx)),
		verax.Field(&r.Age, verax.Required, verax.Min(18), verax.Max(120)),
	)
}

type Planet struct {
	Position int    `json:"position"`
	Name     string `json:"name" solar:"planet_name"`
	Life     float64
}

func (p *Planet) Validate() error {
	return verax.ValidateStruct(
		p,
		verax.Field(&p.Position, verax.Min(1), verax.Max(8)),
		verax.Field(&p.Name, verax.Length(4, 7)).Tag("solar"),
		verax.Field(&p.Life, verax.Min(0.0), verax.Max(1.0)),
	)
}

func ExampleValidate_primitive_int() {
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
}

func ExampleValidateStruct() {
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
}

func ExampleValidateStruct_custom_tag() {
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
}

func ExampleValidator() {
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
}

func ExampleValidator_quick_start() {
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
}

func ExampleValidate_slices() {
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
}

func ExampleValidate_maps() {
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
}

func ExampleMap() {
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
}

func ExampleSet() {
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
}

func ExampleBy() {
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
}

type Range struct {
	Start int
	End   int
}

func ExampleSkip() {
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
}

func ExampleWhen() {
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
}

func ExampleRule_conditioned() {
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
}

type UserDoesNotExistRule struct{}

func (u UserDoesNotExistRule) Validate(v any) error {
	username, err := verax.EnsureString(v)
	if err != nil {
		return verax.ErrInvType
	}

	// Check if the username exists in a database.

	format := "user %s already exist"
	return verax.NewErrorf(format, username, xrr.WithCode("ECMustNotExist"))
}

func ExampleRule() {
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
}

func ExampleRule_customMessage() {
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
}

func ExampleRule_customErrorCode() {
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
}

func ExampleBuilders_encode() {
	rule := verax.Min(18)

	reg := spec.NewRegistry[verax.Rule]()
	reg.RegisterBuilders(verax.Builders())

	spc, _ := rule.Spec()
	data, _ := reg.EncodeSpec(spc)

	fmt.Println(string(data))
	// Output:
	// {"name":"range-rule","args":{"mode":"min","value":{"type":"int","value":18}}}
}

func ExampleBuilders_decode() {
	data := []byte(`{"name":"range-rule","args":{"mode":{"type":"string","value":"min"},"value":{"type":"int","value":18}}}`)

	reg := spec.NewRegistry[verax.Rule]()
	reg.RegisterBuilders(verax.Builders())

	restored, _ := reg.DecodeAndBuild(data)

	err := verax.Validate(15, restored)
	fmt.Println(err)
	// Output:
	// must be greater or equal to 18
}

// nonEmptyWord is a custom RuleFunc used in ExampleBuilders_by.
// Declared as verax.RuleFunc so the type assertion in ByRuleFromSpec succeeds.
var nonEmptyWord verax.RuleFunc = func(v any) error {
	str, err := verax.EnsureString(v)
	if err != nil {
		return verax.ErrInvType
	}
	if len(str) < 3 {
		return verax.NewError("must be at least 3 characters long", "ECShort")
	}
	return nil
}

func ExampleBuilders_by() {
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
}

// PrintJSON marshals value to JSON string.
func PrintJSON(v any) {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("JSON:\n%s\n", string(data))
}

// PrintError formats error message.
func PrintError(err error) {
	var msg strings.Builder
	for line := range strings.SplitSeq(err.Error(), "; ") {
		msg.WriteString("- " + line + "\n")
	}
	fmt.Printf("ERROR:\n\n%s\n", msg.String())
}
