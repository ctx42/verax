package verax_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/verax"
)

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
	// - must be no greater than 44
	//
	// JSON:
	// {
	//     "code": "ECInvThreshold",
	//     "error": "must be no greater than 44"
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
}

func ExampleSet() {
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
}

func ExampleBy() {
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
}

type Range struct {
	Start int
	End   int
}

func ExampleSkip() {
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
}

func ExampleWhen() {
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
}

func ExampleConditioner() {
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
}

type UserDoesNotExistRule struct{}

func (u UserDoesNotExistRule) Validate(v any) error {
	username, err := verax.EnsureString(v)
	if err != nil {
		return verax.ErrInvType
	}

	// Check if the username exists in a database.

	err = fmt.Errorf("user %s already exist", username)
	return xrr.Wrap(err, xrr.WithCode("ECMustNotExist"))
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

func ExampleCustomizer_Error() {
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
}

func ExampleCustomizer_Code() {
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
	var msg string
	for _, line := range strings.Split(err.Error(), "; ") {
		msg += "- " + line + "\n"
	}
	fmt.Printf("ERROR:\n\n%s\n", msg)
}
