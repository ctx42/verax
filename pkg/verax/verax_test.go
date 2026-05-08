// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"regexp"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Check(t *testing.T) {
	t.Run("the adapted function is called", func(t *testing.T) {
		// --- Given ---
		var with int
		fn := func(have int) bool { with = have; return true }

		// --- When ---
		have := Check(fn, "test err", "ECTst")

		// --- Then ---
		err := have(42)

		assert.NoError(t, err)
		assert.Equal(t, 42, with)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- Given ---
		fn := func(have int) bool { return true }

		// --- When ---
		have := Check(fn, "test err", "ECTst")

		// --- Then ---
		err := have(true)
		assert.SameType(t, &InternalError{}, err)
		xrrtest.AssertEqual(t, "test err: expected int, got bool (ECInvType)", err)
		xrrtest.AssertCode(t, ECInvType, err)
	})

	t.Run("error returned when the function returns false", func(t *testing.T) {
		// --- Given ---
		fn := func(have int) bool { return false }

		// --- When ---
		have := Check(fn, "test err", "ECTst")

		// --- Then ---
		err := have(42)
		assert.SameType(t, &Error{}, err)
		xrrtest.AssertEqual(t, "test err (ECTst)", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})
}

func Test_Validate(t *testing.T) {
	t.Run("valid nil no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(iInt)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid pointer no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(pString)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil pointer no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(pStringNil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil slice no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(dSlice)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil array no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(dArray)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil map no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(dMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid one rule", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", Equal("abc"))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid both ok", func(t *testing.T) {
		// --- When ---
		err := Validate(42, Min(42), Equal(42))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid first ok skip second", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", Equal("abc"), Skip, Fail("test err", "ECTst"))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid first ok skip when true second", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", Equal("abc"), Skip.When(true), Equal("xyz"))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid struct with validator", func(t *testing.T) {
		// --- Given ---
		s := &ModelPtr{"abc"}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid struct with validator", func(t *testing.T) {
		// --- Given ---
		s := &ModelPtr{"xyz"}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		wMsg := "FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid first fails", func(t *testing.T) {
		// --- When ---
		err := Validate("123", Equal("abc"), Equal("xyz"))

		// --- Then ---
		wMsg := "must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid second fails", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", Equal("abc"), Equal("xyz"))

		// --- Then ---
		wMsg := "must be equal to 'xyz' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid first fails skip second", func(t *testing.T) {
		// --- When ---
		err := Validate("123", Equal("abc"), Skip, Fail("test err", "ECTst"))

		// --- Then ---
		wMsg := "must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid first fails skip when true second", func(t *testing.T) {
		// --- When ---
		err := Validate(
			"123",
			Equal("abc"),
			Skip.When(true),
			Fail("test err", "ECTst"),
		)

		// --- Then ---
		wMsg := "must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid first fail skip when false second", func(t *testing.T) {
		// --- When ---
		err := Validate(
			"123",
			Equal("abc"),
			Skip.When(false),
			Equal("xyz"),
		)

		// --- Then ---
		wMsg := "must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid first ok skip when false second", func(t *testing.T) {
		// --- When ---
		err := Validate(
			"abc",
			Equal("abc"),
			Skip.When(false),
			Fail("test err", "ECTst"),
		)

		// --- Then ---
		xrrtest.AssertEqual(t, "test err (ECTst)", err)
	})

	t.Run("invalid many in slice", func(t *testing.T) {
		// --- Given ---
		s := []*ModelPtr{{"xyz"}, {"xyz"}}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		want := "" +
			"0.FStr: must be equal to 'abc' (ECNotEqual); " +
			"1.FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, want, err)
	})

	t.Run("valid all in slice", func(t *testing.T) {
		// --- Given ---
		s := []*ModelPtr{{"abc"}, {"abc"}}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid many in map", func(t *testing.T) {
		// --- Given ---
		m := map[int]*ModelPtr{0: {"xyz"}, 4: {"xyz"}}

		// --- When ---
		err := Validate(m)

		// --- Then ---
		want := "" +
			"0.FStr: must be equal to 'abc' (ECNotEqual); " +
			"4.FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, want, err)
	})

	t.Run("valid all in map", func(t *testing.T) {
		// --- Given ---
		m := map[int]*ModelPtr{0: {"abc"}, 4: {"abc"}}

		// --- When ---
		err := Validate(m)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("ValidateWith success", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{"111"}

		// --- When ---
		err := Validate(m, Equal("111"), NotNil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("ValidateWith runs provided rule", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{value: "abc"}

		// --- When ---
		err := Validate(m, Equal("111"))

		// --- Then ---
		assert.Error(t, err)
		assert.ErrorEqual(t, "must be equal to '111'", err)
	})

	t.Run("ValidateWith additional validation in the type", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{"too_long"}

		// --- When ---
		err := Validate(m, Equal("wrong_value"))

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
	})

	t.Run("ValidateWith multiple rules", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{"too_long"}

		// --- When ---
		err := Validate(m, Equal("wrong_value"))

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
	})
}

func Test_ValidateName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := ValidateNamed("field", 42, Equal(42))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		// --- When ---
		err := ValidateNamed("field", 43, Equal(42))

		// --- Then ---
		wMsg := "field: must be equal to '42' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})
}

func Test_Set(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		// --- Given ---
		r := Set{
			Min(40),
			Max(45),
		}

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		// --- Given ---
		r := Set{
			Min(40),
			Max(45),
		}

		// --- When ---
		err := r.Validate(39)

		// --- Then ---
		assert.ErrorEqual(t, "must be greater or equal to 40", err)
		xrrtest.AssertCode(t, ECInvRange, err)
	})
}

func Test_Set_Spec(t *testing.T) {
	t.Run("no rules", func(t *testing.T) {
		// --- Given ---
		r := Set{}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, SetRuleName, have.Name)
		assert.Equal(t, map[string]any{}, have.Args)
	})

	t.Run("with rules", func(t *testing.T) {
		// --- Given ---
		rules := []Rule{Min(42), Max(44)}
		r := Set(rules)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, SetRuleName, have.Name)
		wArgs := map[string]any{spec.ArgTypes: rules}
		assert.Equal(t, wArgs, have.Args)
		assert.NotSame(t, rules, have.Args[spec.ArgTypes])
	})
}

func Test_SetRuleFromSpec(t *testing.T) {
	t.Run("error - wrong spec name", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := SetRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `set-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Nil(t, have)
	})

	t.Run("error - missing types argument", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(SetRuleName)

		// --- When ---
		have, err := SetRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "set-rule: spec missing required argument: types"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Nil(t, have)
	})

	t.Run("success - with rules", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(SetRuleName).
			SetArg(spec.ArgTypes, []Rule{Min(42), Max(44)})

		// --- When ---
		have, err := SetRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Set{Min(42), Max(44)}, have)
	})
}

func Test_Set_Spec_SetRuleFromSpec_round_trip(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		// --- Given ---
		want := Set{Min(42), Required}
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := SetRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}

func Test_Named_Set_Get_GetOrNoop(t *testing.T) {
	t.Run("get set", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		g1 := nr.Get("r1")
		assert.NoError(t, g1.Validate(1))

		g2 := nr.Get("r2")
		assert.NoError(t, g2.Validate(2))
	})

	t.Run("get not present rule", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		assert.Nil(t, nr.Get("r3"))
	})

	t.Run("get safe existing", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		g2 := nr.GetOrNoop("r2")
		assert.NotNil(t, g2)
		assert.NoError(t, g2.Validate(2))
	})

	t.Run("get safe not existing", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		g3 := nr.GetOrNoop("r3")
		assert.NotNil(t, g3)
		assert.NoError(t, g3.Validate(3))
	})

	t.Run("set overrides existing", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r1", r2)

		// --- Then ---
		g1 := nr.Get("r1")
		assert.Error(t, g1.Validate(1))
		assert.NoError(t, g1.Validate(2))
	})
}

func Test_Named_GetOrError(t *testing.T) {
	t.Run("rule exists", func(t *testing.T) {
		// --- Given ---
		r1 := In(1, 11)
		r2 := In(2, 22)

		nr := NewNamed().Set("r1", r1).Set("r2", r2)

		// --- When ---
		have := nr.GetOrError("r1")

		// --- Then ---
		assert.Equal(t, r1, have)
	})

	t.Run("rule does not exist", func(t *testing.T) {
		// --- Given ---
		r1 := In(1, 11)
		r2 := In(2, 22)

		nr := NewNamed().Set("r1", r1).Set("r2", r2)

		// --- When ---
		have := nr.GetOrError("r3")

		// --- Then ---
		err := have.Validate(11)
		xrrtest.AssertEqual(t, "unknown rule (ECUnkRule)", err)
	})
}

func Test_Rule_encoding_round_trip_tabular(t *testing.T) {
	ruleFunc := RuleFunc(func(v any) error { return nil })
	ruleFuncSrc := must.Value(spec.NewSource("ruleFunc", ruleFunc))

	tt := []struct {
		testN string

		src SpecableRule
		bld spec.Builder[Rule]
	}{
		// AbsentRule.

		{
			"AbsentRule-Nil with error message and code",
			Nil.Message("test msg").Code("ECTst"),
			AsRuleBuilder(AbsentRuleFromSpec),
		},
		{
			"AbsentRule-Empty with error message and code",
			Empty.Message("test msg").Code("ECTst"),
			AsRuleBuilder(AbsentRuleFromSpec),
		},

		// ByRule.

		{
			"ByRule with error message and code",
			By(ruleFunc).Message("test msg").Code("ECTst"),
			AsRuleBuilder(ByRuleFromSpec),
		},

		// ContainRule.

		{
			"ContainRule using Equal with error message and code",
			Contain(Equal(42)).Message("test msg").Code("ECTst"),
			AsRuleBuilder(ContainRuleFromSpec),
		},
		{
			"ContainRule using NotEqual with error message and code",
			Contain(NotEqual(42)).Message("test msg").Code("ECTst"),
			AsRuleBuilder(ContainRuleFromSpec),
		},

		// EachRule.

		{
			"EachRule",
			Each(Min(42), Max(44)),
			AsRuleBuilder(EachRuleFromSpec),
		},

		// EqualRule.

		{
			"EqualRule-Equal - with error message and code",
			Equal(42).Message("test msg").Code("ECTst"),
			AsRuleBuilder(EqualRuleFromSpec),
		},
		{
			"EqualRule-NotEqual - with error message and code",
			NotEqual(42).Message("test msg").Code("ECTst"),
			AsRuleBuilder(EqualRuleFromSpec),
		},
		{
			"EqualRule-EqualField",
			EqualField(42, "field"),
			AsRuleBuilder(EqualRuleFromSpec),
		},
		{
			"EqualRule-EqualField - with error message and code",
			EqualField(42, "field").Message("test msg").Code("ECTst"),
			AsRuleBuilder(EqualRuleFromSpec),
		},
		{
			"EqualRule-NotEqualField",
			NotEqualField(42, "field"),
			AsRuleBuilder(EqualRuleFromSpec),
		},
		{
			"EqualRule-NotEqualField - with error message and code",
			NotEqualField(42, "field").Message("test msg").Code("ECTst"),
			AsRuleBuilder(EqualRuleFromSpec),
		},

		// FailRule.

		{
			"FailRule",
			Fail("test msg", "ECTst"),
			AsRuleBuilder(FailRuleFromSpec),
		},

		// InRule.

		{
			"InRule - with error message and code",
			In(1, 2, 3).Message("test msg").Code("ECTst"),
			AsRuleBuilder(InRuleFromSpec),
		},

		// LengthRule.

		{
			"LengthRule-Length - with error message and code",
			Length(42, 44).Message("test msg").Code("ECTst"),
			AsRuleBuilder(LengthRuleFromSpec),
		},
		{
			"LengthRule-RuneLength - with error message and code",
			RuneLength(42, 44).Message("test msg").Code("ECTst"),
			AsRuleBuilder(LengthRuleFromSpec),
		},

		// MapRule.

		{
			"MapRule",
			Map(Key(1, Min(42)), Key(3, Max(44))),
			AsRuleBuilder(MapRuleFromSpec),
		},

		// MatchRule.

		{
			"MatchRule - with error message and code",
			Match(regexp.MustCompile(`\d+`)).Message("test msg").Code("ECTst"),
			AsRuleBuilder(MatchRuleFromSpec),
		},

		// NoopRule.

		{
			"NoopRule",
			Noop,
			AsRuleBuilder(NoopRuleFromSpec),
		},

		// RequiredRule.

		{
			"RequiredRule-Required - with error message and code",
			Required.Message("test msg").Code("ECTst"),
			AsRuleBuilder(RequiredRuleFromSpec),
		},
		{
			"RequiredRule-NotEmpty - with error message and code",
			NotEmpty.Message("test msg").Code("ECTst"),
			AsRuleBuilder(RequiredRuleFromSpec),
		},
		{
			"RequiredRule-NotNil - with error message and code",
			NotNil.Message("test msg").Code("ECTst"),
			AsRuleBuilder(RequiredRuleFromSpec),
		},

		// SkipRule.

		{
			"SkipRule",
			Skip,
			AsRuleBuilder(SkipRuleFromSpec),
		},

		// RangeRule.

		{
			"RangeRule-Min - with error message and code",
			Min(42).Exclusive().Message("test msg").Code("ECTst"),
			AsRuleBuilder(RangeRuleFromSpec),
		},
		{
			"RangeRule-Max - with error message and code",
			Max(42).Exclusive().Message("test msg").Code("ECTst"),
			AsRuleBuilder(RangeRuleFromSpec),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// Setup Registry.
			reg := spec.NewRegistry[Rule]()
			reg.RegisterSource(ruleFuncSrc)
			reg.RegisterBuilders(Builders())

			// Get spec.
			srcSpc, err := tc.src.Spec()
			assert.NoError(t, err)

			// Encode.
			data, err := reg.EncodeSpec(srcSpc)
			assert.NoError(t, err)

			// Decode.
			dstSpc := &spec.Spec{}
			assert.NoError(t, reg.DecodeSpec(data, dstSpc))

			// Build.
			dst, err := tc.bld(dstSpc)
			assert.NoError(t, err)

			// Compare.
			assert.Equal(t, tc.src, dst)
		})
	}
}
