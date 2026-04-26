// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Each(t *testing.T) {
	// --- Given ---
	r0 := TstRule{"r0"}
	r1 := TstRule{"r1"}

	// --- When ---
	have := Each(r0, r1)

	// --- Then ---
	want := EachRule{
		condition: true,
		rules:     []Rule{r0, r1},
	}
	assert.Equal(t, want, have)
}

func Test_EachRule_Validate(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := EachRule{condition: false, rules: []Rule{Equal("abc")}}

		// --- When ---
		err := r.Validate([]string{"xyz"})

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - slice of struct pointers", func(t *testing.T) {
		// --- Given ---
		var s []*TStruct
		ts0 := NewTStruct()
		ts1 := NewTStruct()
		s = append(s, &ts0, &ts1)
		s[1].FStr = "wrong"

		fn := func(v any) error {
			if v.(*TStruct).FStr != "FStr" {
				return NewError("error", "ECTst")
			}
			return nil
		}

		// --- When ---
		err := Each(By(fn)).Validate(s)

		// --- Then ---
		assert.SameType(t, &FieldsError{}, err)
		xrrtest.AssertEqual(t, "1: error (ECTst)", err)
	})

	t.Run("error - slice of struct values", func(t *testing.T) {
		// --- Given ---
		s := []TStruct{NewTStruct(), NewTStruct()}
		s[1].FStr = "wrong"

		fn := func(v any) error {
			if v.(TStruct).FStr != "FStr" {
				return NewError("error", "ECTst")
			}
			return nil
		}

		// --- When ---
		err := Each(By(fn)).Validate(s)

		// --- Then ---
		assert.SameType(t, &FieldsError{}, err)
		xrrtest.AssertEqual(t, "1: error (ECTst)", err)
	})

	t.Run("error - not iterable value", func(t *testing.T) {
		// --- Given ---
		r := Each(TstRule{})

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		xrrtest.AssertEqual(t, "must be iterable (ECInvType)", err)
	})
}

func Test_EachRule_Validate_valid_tabular(t *testing.T) {
	nonEmptyChan := make(chan int, 1)
	nonEmptyChan <- 1

	tt := []struct {
		testN string

		rules []Rule
		have  any
	}{
		{
			"slice empty",
			[]Rule{},
			[]string{},
		},
		{
			"slice with values",
			[]Rule{Required},
			[]string{"abc", "def"},
		},
		{
			"slice with validators",
			[]Rule{Required},
			[]ModelVal{{"abc"}, {"abc"}},
		},
		{
			"map empty",
			[]Rule{},
			map[string]string{},
		},
		{
			"map with keys",
			[]Rule{Required},
			map[string]string{"key0": "val0", "key1": "val1"},
		},
		{
			"map with validator keys",
			[]Rule{Required},
			map[string]ModelVal{"key0": {"abc"}, "key1": {"abc"}},
		},
		{
			"array empty",
			[]Rule{},
			[...]string{},
		},
		{
			"array with values",
			[]Rule{Required},
			[...]string{"abc", "def"},
		},
		{
			"array with validators",
			[]Rule{Required},
			[...]ModelVal{{"abc"}, {"abc"}},
		},
		{
			"channel instances",
			[]Rule{Required},
			[]chan int{nonEmptyChan},
		},
		{
			"function instances",
			[]Rule{Required},
			[]func(any) bool{iFunc},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Each(tc.rules...)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_EachRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rules []Rule
		have  any
		err   string
	}{
		{
			"slice with values",
			[]Rule{Required},
			[]string{"def", ""},
			"1: cannot be blank (ECRequired)",
		},
		{
			"slice with nils",
			[]Rule{Required},
			[]*string{pString, pStringNil},
			"1: cannot be blank (ECRequired)",
		},
		{
			"slice with validators",
			[]Rule{Required},
			[]ModelVal{{"abc"}, {"def"}},
			"1.FStr: must be equal to 'abc' (ECNotEqual)",
		},
		{
			"map with keys",
			[]Rule{Required},
			map[string]string{"key0": "val0", "key1": ""},
			"key1: cannot be blank (ECRequired)",
		},
		{
			"map with nils",
			[]Rule{Required},
			map[string]*string{"key0": pString, "key1": pStringNil},
			"key1: cannot be blank (ECRequired)",
		},
		{
			"map with validator keys",
			[]Rule{Required},
			map[string]ModelVal{"key0": {"abc"}, "key1": {"def"}},
			"key1.FStr: must be equal to 'abc' (ECNotEqual)",
		},
		{
			"array with values",
			[]Rule{Required},
			[...]string{"abc", ""},
			"1: cannot be blank (ECRequired)",
		},
		{
			"array with validators",
			[]Rule{Required},
			[...]ModelVal{{"abc"}, {"def"}},
			"1.FStr: must be equal to 'abc' (ECNotEqual)",
		},
		{
			"array with nils",
			[]Rule{Required},
			[...]*string{pString, nil},
			"1: cannot be blank (ECRequired)",
		},
		{
			"channel declared",
			[]Rule{Required},
			[]any{dChan},
			"0: cannot be blank (ECRequired)",
		},
		{
			"channel zero length",
			[]Rule{Required},
			[]chan int{iChan},
			"0: cannot be blank (ECRequired)",
		},
		{
			"function declared",
			[]Rule{Required},
			[]any{dFunc},
			"0: cannot be blank (ECRequired)",
		},
		{
			"pointers",
			[]Rule{Required},
			[]*int{pIntNil, nil},
			"0: cannot be blank (ECRequired); 1: cannot be blank (ECRequired)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Each(tc.rules...)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &FieldsError{}, err)
			xrrtest.AssertEqual(t, tc.err, err)
		})
	}
}

func Test_EachRule_When(t *testing.T) {
	// --- Given ---
	r := EachRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_EachRule_Spec(t *testing.T) {
	t.Run("no rules", func(t *testing.T) {
		// --- Given ---
		r := Each()

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EachRuleName, have.Name)
		assert.Equal(t, map[string]any{}, have.Args)
	})

	t.Run("with rules", func(t *testing.T) {
		// --- Given ---
		rules := []Rule{Min(42).Exclusive(), Max(44)}
		r := Each(rules...)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EachRuleName, have.Name)
		wArgs := map[string]any{spec.ArgTypes: rules}
		assert.Equal(t, wArgs, have.Args)
		assert.NotSame(t, rules, have.Args[spec.ArgTypes])
	})
}

func Test_EachRuleFromSpec(t *testing.T) {
	t.Run("error - wrong spec name", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := EachRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `each-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - missing rules argument", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EachRuleName)

		// --- When ---
		have, err := EachRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "each-rule: spec missing required argument: types"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("success - with rules", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EachRuleName).
			SetArg(spec.ArgTypes, []Rule{Min(42).Exclusive(), Max(44)})

		// --- When ---
		have, err := EachRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Each(Min(42).Exclusive(), Max(44)), have)
	})

	t.Run("success - with a custom message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EachRuleName).
			SetArg(spec.ArgTypes, []Rule{Required.Message("custom msg")})

		// --- When ---
		have, err := EachRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Each(Required.Message("custom msg")), have)
	})

	t.Run("success - with custom code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EachRuleName).
			SetArg(spec.ArgTypes, []Rule{Required.Code("ECTst")})

		// --- When ---
		have, err := EachRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Each(Required.Code("ECTst")), have)
	})
}

func Test_EachRule_Spec_ContainRuleFromSpec_round_trip(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		// --- Given ---
		want := Each(Min(42), Required)
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := EachRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
