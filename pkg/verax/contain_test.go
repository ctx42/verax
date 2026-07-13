// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Contain(t *testing.T) {
	// --- Given ---
	eq := Equal(42)

	// --- When ---
	have := Contain(eq)

	// --- Then ---
	assert.Equal(t, eq, have.rule)
}

func Test_ContainRule_Validate(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := Contain(Equal(42)).When(false)

		// --- When ---
		err := r.Validate([]int{44})

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip validation when the condition is false and empty", func(t *testing.T) {
		// --- Given ---
		r := Contain(Equal(42)).When(false)

		// --- When ---
		err := r.Validate([]int{})

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - not iterable value", func(t *testing.T) {
		// --- Given ---
		r := Contain(Equal(42))

		// --- When ---
		err := r.Validate("C")

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		xrrtest.AssertEqual(t, "must be iterable (ECInvType)", err)
	})
}

func Test_ContainRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule EqualRule
		have any
	}{
		{"slice of int", Equal(2), []int{1, 2, 3}},
		{"slice of string", Equal("c"), []string{"a", "b", "c"}},
		{"array of int", Equal(2), [...]int{1, 2, 3}},
		{"array of string", Equal("c"), [...]string{"a", "b", "c"}},

		{"map string:int", Equal(2), map[string]int{"A": 1, "B": 2, "C": 3}},
		{"map int:string", Equal("C"), map[int]string{1: "A", 2: "B", 3: "C"}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Contain(tc.rule)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_ContainRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule EqualRule
		have any
		err  string
		code string
	}{
		{
			"slice does not contain",
			Equal(0),
			[]int{1, 2, 3},
			"must contain at least one '0' value",
			ECNotEqual,
		},
		{
			"empty slice",
			Equal(0),
			[]int{},
			"must contain at least one '0' value",
			ECNotEqual,
		},
		{
			"nil slice",
			Equal(0),
			[]int(nil),
			"must contain at least one '0' value",
			ECNotEqual,
		},
		{
			"array does not contain",
			Equal(4),
			[...]int{1, 2, 3},
			"must contain at least one '4' value",
			ECNotEqual,
		},
		{
			"empty array",
			Equal(4),
			[...]int{},
			"must contain at least one '4' value",
			ECNotEqual,
		},
		{
			"map does not contain",
			Equal("D"),
			map[string]int{"A": 1, "B": 2, "C": 3},
			"must contain at least one 'D' value",
			ECNotEqual,
		},
		{
			"empty map",
			Equal("D"),
			map[string]int{},
			"must contain at least one 'D' value", ECNotEqual,
		},
		{
			"nil map",
			Equal("D"),
			map[string]int(nil),
			"must contain at least one 'D' value",
			ECNotEqual,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Contain(tc.rule)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_ContainRule_When(t *testing.T) {
	// --- Given ---
	r := ContainRule{rule: EqualRule{}}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.rule.condition)
}

func Test_ContainRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := ContainRule{rule: EqualRule{}}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.rule.msg)
		assert.Equal(t, flgCustomMsg, have.rule.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := ContainRule{rule: EqualRule{msg: "test err"}}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.rule.msg)
		assert.Zero(t, have.rule.flags)
	})
}

func Test_ContainRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := ContainRule{rule: EqualRule{}}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.rule.code)
		assert.Equal(t, flgCustomCode, have.rule.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := ContainRule{rule: EqualRule{code: "ECTst"}}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.rule.code)
		assert.Zero(t, have.rule.flags)
	})
}

func Test_ContainRule_Spec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		eq := Equal("abc")
		r := Contain(eq)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, ContainRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "equal", spec.ArgValue: "abc"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		eq := Equal(func() {})
		r := Contain(eq)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, "equal-rule(equal): template render error", err)
		xrrtest.AssertCode(t, ECInternal, err)
		assert.Nil(t, have)
	})

	t.Run("Contain - JSON encode", func(t *testing.T) {
		// --- Contain ---
		reg := spec.NewRegistry[Rule]()

		// --- When ---
		have, err := Contain(Equal("foo")).Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(reg.EncodeSpec(have))
		want := `{
			"name": "contain-rule",
			"args": {
				"mode": "equal",
				"value": "foo"
			}
		}`
		assert.JSON(t, want, data)
	})

	t.Run("Contain - JSON decode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		reg.RegisterBuilders(Builders())
		data := []byte(`{
			"name": "contain-rule",
			"args": {
				"mode": "equal",
				"value": "foo"
			}
		}`)

		// --- When ---
		have, err := reg.DecodeAndBuild(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Contain(Equal("foo")), have)
	})
}

func Test_ContainRuleFromSpec(t *testing.T) {
	t.Run("error - not contain rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := ContainRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `contain-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(ContainRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := ContainRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Contain(Equal(42))
		assert.Equal(t, wRule, have)
	})

	t.Run("error - mode argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(ContainRuleName)

		// --- When ---
		have, err := ContainRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "contain-rule: spec missing required argument: mode"
		assert.ErrorEqual(t, wMsg, err)
		assert.Zero(t, have)
	})
}

func Test_ContainRule_Spec_ContainRuleFromSpec_round_trip(t *testing.T) {
	t.Run("Contain - Equal - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Contain(Equal(42)).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := ContainRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("Contain - with EqualFunc message and code", func(t *testing.T) {
		// --- Given ---
		fn := EqualFunc(func(any, any) error { return nil })
		want := Contain(Equal(42).With(fn)).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := ContainRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
