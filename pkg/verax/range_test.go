// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/internal/test"
	"github.com/ctx42/verax/pkg/spec"
)

func Test_Min(t *testing.T) {
	t.Run("supported type", func(t *testing.T) {
		// --- When ---
		have := Min(42)

		// --- Then ---
		want := RangeRule{
			mode:      "min",
			threshold: 42,
			fn:        compareInt,
			condition: true,
			tpl:       msgGreaterOrEqual,
			msg:       "must be greater or equal to 42",
			code:      ECInvRange,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - not supported type", func(t *testing.T) {
		// --- Given ---
		val := func() {}

		// --- When ---
		have := Min(val)

		// --- Then ---
		want := RangeRule{
			mode:      "min",
			threshold: val,
			fn:        nil,
			condition: true,
			tpl:       msgGreaterOrEqual,
			msg:       "",
			code:      ECInvRange,
			flags:     0,
		}
		assert.Equal(t, want, have, check.WithSkipTrail("RangeRule.sticky"))
		assert.SameType(t, &InternalError{}, have.sticky)
		xrrtest.AssertCode(t, ECInvType, have.sticky)
	})
}

func Test_Max(t *testing.T) {
	t.Run("supported type", func(t *testing.T) {
		// --- When ---
		have := Max(42)

		// --- Then ---
		want := RangeRule{
			mode:      "max",
			threshold: 42,
			fn:        compareInt,
			condition: true,
			tpl:       msgLessOrEqual,
			msg:       "must be less or equal to 42",
			code:      ECInvRange,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)

		assert.Equal(t, "max", have.mode)
		assert.Equal(t, 42, have.threshold)
		assert.Same(t, compareInt, have.fn)
		assert.True(t, have.condition)
		assert.Equal(t, msgLessOrEqual, have.tpl)
		assert.Equal(t, "must be less or equal to 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.NoError(t, have.sticky)
		assert.Zero(t, have.flags)
	})

	t.Run("error - not supported type", func(t *testing.T) {
		// --- Given ---
		val := func() {}

		// --- When ---
		have := Max(val)

		// --- Then ---
		want := RangeRule{
			mode:      "max",
			threshold: val,
			fn:        nil,
			condition: true,
			tpl:       msgLessOrEqual,
			msg:       "",
			code:      ECInvRange,
			flags:     0,
		}
		assert.Equal(t, want, have, check.WithSkipTrail("RangeRule.sticky"))
		assert.SameType(t, &InternalError{}, have.sticky)
		xrrtest.AssertCode(t, ECInvType, have.sticky)
	})
}

func Test_RangeRule_Exclusive_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule RangeRule
		wTpl string
		wMsg string
	}{
		{
			"greater or equal",
			Min(42),
			msgGreaterThan,
			"must be greater than 42",
		},
		{
			"less or equal",
			Max(42),
			msgLessThan,
			"must be less than 42",
		},
		{
			"greater than - noop",
			Min(42).Exclusive(),
			msgGreaterThan,
			"must be greater than 42",
		},
		{
			"less than - noop",
			Max(42).Exclusive(),
			msgLessThan,
			"must be less than 42",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := tc.rule

			// --- When ---
			have := r.Exclusive()

			// --- Then ---
			assert.Equal(t, tc.wTpl, have.tpl)
			assert.Equal(t, tc.wMsg, have.msg)
			assert.NoError(t, have.sticky)
		})
	}
}

func Test_RangeRule_Exclusive(t *testing.T) {
	t.Run("when the sticky error is not nil", func(t *testing.T) {
		// --- Given ---
		r := Max(42)
		r.sticky = ErrTst

		// --- When ---
		have := r.Exclusive()

		// --- Then ---
		assert.Equal(t, msgLessOrEqual, have.tpl)
		assert.Equal(t, "must be less or equal to 42", have.msg)
		assert.Equal(t, "max", r.mode)
	})
}

func Test_RangeRule_With(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		fn := func(want, have any) (int, error) { return 0, nil }
		r := RangeRule{}

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Same(t, fn, have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})

	t.Run("ignores ECInvType sticky so caller can provide a custom function", func(t *testing.T) {
		// --- Given ---
		fn := func(want, have any) (int, error) { return 0, nil }
		r := Min(func() {})

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Same(t, fn, have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})

	t.Run("changes nothing when a sticky error is set", func(t *testing.T) {
		// --- Given ---
		fn := func(want, have any) (int, error) { return 0, nil }
		r := RangeRule{sticky: errors.New("test error")}

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Nil(t, have.fn)
		assert.Zero(t, have.flags)
	})
}

func Test_RangeRule_Validate(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := RangeRule{sticky: ErrTst}

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := Max(1).When(false)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil ok", func(t *testing.T) {
		// --- Given ---
		r := Min(42)

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty ok", func(t *testing.T) {
		// --- Given ---
		r := Min(42)

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- Given ---
		r := Min(44)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "must be greater or equal to 44", err)
		xrrtest.AssertCode(t, ECInvRange, err)
	})

	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := Min(42)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("custom fn error is handled unchanged", func(t *testing.T) {
		// --- Given ---
		var w, h any
		fn := func(want, have any) (int, error) {
			w = want
			h = have
			return -1, ErrTst
		}
		r := Max(42).With(fn)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.Same(t, ErrTst, err)
		assert.Equal(t, 42, w)
		assert.Equal(t, 44, h)
	})

	t.Run("error - custom fn error replaced by custom msg and code", func(t *testing.T) {
		// --- Given ---
		fn := func(_, _ any) (int, error) { return 0, ErrTst }
		r := Min(44).With(fn).Message("custom {{.value}}").Code("ECTst")

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "custom 44", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - custom fn error with a custom message", func(t *testing.T) {
		// --- Given ---
		fn := func(_, _ any) (int, error) { return 0, ErrTst }
		r := Min(42).With(fn).Message("custom msg")

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "custom msg", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - custom fn error with custom code", func(t *testing.T) {
		// --- Given ---
		fn := func(_, _ any) (int, error) { return 0, ErrTst }
		r := Min(42).With(fn).Code("ECTst")

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorIs(t, ErrTst, err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - not supported type", func(t *testing.T) {
		// --- Given ---
		type Int int
		r := Max(42)

		// --- When ---
		err := r.Validate(Int(42))

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert verax.Int to int64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
	})
}

func Test_RangeRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule RangeRule
		have any
	}{
		{"min nil", Min(42), nil},
		{"max nil", Max(42), nil},
		{"min empty string", Min(42), ""},
		{"max empty string", Max(42), ""},

		{"min", Min(42), 42},
		{"min exclusive", Min(42).Exclusive(), 43},
		{"max", Max(42), 42},
		{"max exclusive", Max(42).Exclusive(), 41},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := tc.rule

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_RangeRule_Validate_invalid_tabular(t *testing.T) {
	cmp := func(_, _ any) (int, error) { return 0, ErrTst }

	tt := []struct {
		testN string

		rule       RangeRule
		have       any
		customMsg  string
		customCode string
		error      string
	}{
		{
			"custom fn",
			Min(42).With(cmp),
			42,
			"",
			"",
			"test msg (ECTst)",
		},
		{
			"custom fn with custom msg",
			Min(42).With(cmp),
			42,
			"test msg",
			"",
			"test msg (ECTst)",
		},
		{
			"custom fn with custom code",
			Min(42).With(cmp),
			42,
			"",
			"MyCode",
			"test msg (MyCode)",
		},
		{
			"custom fn with msg and code",
			Min(42).With(cmp),
			42,
			"custom msg",
			"MyCode",
			"custom msg (MyCode)",
		},
		{
			"outcome",
			Max(41),
			42,
			"",
			"",
			"must be less or equal to 41 (ECInvRange)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := tc.rule.Message(tc.customMsg).Code(tc.customCode)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			xrrtest.AssertEqual(t, tc.error, err)
		})
	}
}

func Test_RangeRule_When(t *testing.T) {
	// --- Given ---
	r := RangeRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_RangeRule_Message(t *testing.T) {
	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, msgLessOrEqual, have.tpl)
		assert.Equal(t, "must be less or equal to 42", have.msg)
		assert.Zero(t, have.flags)
	})

	t.Run("when the sticky error is not nil", func(t *testing.T) {
		// --- Given ---
		r := Max(42)
		r.sticky = ErrTst

		// --- When ---
		have := r.Message("{{.value}}")

		// --- Then ---
		assert.Equal(t, msgLessOrEqual, have.tpl)
		assert.Equal(t, "must be less or equal to 42", have.msg)
		assert.Zero(t, have.flags)
	})

	t.Run("valid template", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		have := r.Message("tpl {{.value}}")

		// --- Then ---
		assert.NoError(t, have.sticky)
		assert.Equal(t, "tpl {{.value}}", have.tpl)
		assert.Equal(t, "tpl 42", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("error - invalid template", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		have := r.Message("bad {{.")

		// --- Then ---
		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "range-rule(max): custom template parse error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
		assert.Equal(t, "must be less or equal to {{.value}}", have.tpl)
		assert.Equal(t, "must be less or equal to 42", have.msg)
		assert.Zero(t, have.flags)
	})

	t.Run("error - with not supported custom placeholder", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		have := r.Message("{{.custom}}")

		// --- Then ---
		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "range-rule(max): custom template render error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
		assert.Equal(t, "must be less or equal to {{.value}}", have.tpl)
		assert.Equal(t, "must be less or equal to 42", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_RangeRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := RangeRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := RangeRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_RangeRule_Spec(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := RangeRule{sticky: ErrTst}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.Same(t, ErrTst, err)
		assert.Nil(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		r := RangeRule{mode: "unknown"}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `range-rule: invalid rule mode: "unknown"`, err)
		xrrtest.AssertCode(t, ECInvRuleMode, err)
		assert.Nil(t, have)
	})

	t.Run("mode min", func(t *testing.T) {
		// --- Given ---
		r := Min(42)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RangeRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "min", spec.ArgValue: 42}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode min-exclusive", func(t *testing.T) {
		// --- Given ---
		r := Min(42).Exclusive()

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RangeRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:       "min-exclusive",
			spec.ArgValue: 42,
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode max", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RangeRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "max", spec.ArgValue: 42}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode max-exclusive", func(t *testing.T) {
		// --- Given ---
		r := Max(42).Exclusive()

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RangeRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:       "max-exclusive",
			spec.ArgValue: 42,
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("with custom error code", func(t *testing.T) {
		// --- Given ---
		r := Min(42).Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		wArgs := map[string]any{
			ArgMode:       "min",
			spec.ArgValue: 42,
			ArgErrCode:    "ECTst",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("with custom error message and code", func(t *testing.T) {
		// --- Given ---
		r := Max(42).Message("{{.value}}").Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		wArgs := map[string]any{
			ArgMode:       "max",
			spec.ArgValue: 42,
			ArgErrMsg:     "{{.value}}",
			ArgErrCode:    "ECTst",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("with a custom cmp function", func(t *testing.T) {
		// --- Given ---
		fn := CompareFunc(func(want, have any) (int, error) { return 0, nil })
		r := Max(42).With(fn)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		wArgs := map[string]any{
			ArgMode:       "max",
			spec.ArgValue: 42,
			spec.ArgSrc:   fn,
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("Min - JSON representation", func(t *testing.T) {
		// --- When ---
		have, err := Min(18).Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(json.Marshal(have))
		want := `{
			"name": "range-rule",
			"args": {
				"mode": "min",
				"value": 18
			}
		}`
		assert.JSON(t, want, data)
	})
}

func Test_RangeRuleFromSpec(t *testing.T) {
	t.Run("error - not range rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `range-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - mode argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---

		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: spec missing required argument: mode"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - want argument is missing", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).SetArg(ArgMode, "min")

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: spec missing required argument: value"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "bad-mode").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `range-rule: invalid spec rule mode: "bad-mode"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - error instantiating", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, true)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: unsupported type comparison: bool"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Zero(t, have)
	})

	t.Run("min", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "min", have.mode)
		assert.Equal(t, 42, have.threshold)
		assert.Same(t, compareInt, have.fn)
		assert.True(t, have.condition)
		assert.Equal(t, msgGreaterOrEqual, have.tpl)
		assert.Equal(t, "must be greater or equal to 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.NoError(t, have.sticky)
		assert.Zero(t, have.flags)
	})

	t.Run("min-exclusive", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min-exclusive").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "min-exclusive", have.mode)
		assert.Equal(t, 42, have.threshold)
		assert.Same(t, compareInt, have.fn)
		assert.True(t, have.condition)
		assert.Equal(t, msgGreaterThan, have.tpl)
		assert.Equal(t, "must be greater than 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.NoError(t, have.sticky)
		assert.Zero(t, have.flags)
	})

	t.Run("max", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "max").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "max", have.mode)
		assert.Equal(t, 42, have.threshold)
		assert.Same(t, compareInt, have.fn)
		assert.True(t, have.condition)
		assert.Equal(t, msgLessOrEqual, have.tpl)
		assert.Equal(t, "must be less or equal to 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.NoError(t, have.sticky)
		assert.Zero(t, have.flags)
	})

	t.Run("max-exclusive", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "max-exclusive").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "max-exclusive", have.mode)
		assert.Equal(t, 42, have.threshold)
		assert.Same(t, compareInt, have.fn)
		assert.True(t, have.condition)
		assert.Equal(t, msgLessThan, have.tpl)
		assert.Equal(t, "must be less than 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.NoError(t, have.sticky)
		assert.Zero(t, have.flags)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrMsg, "test {{.value}}")

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.SameType(t, RangeRule{}, have)
		assert.Equal(t, "test {{.value}}", have.tpl)
		assert.Equal(t, "test 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrMsg, true)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: spec argument \"err_msg\" must be string, got bool"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, msgGreaterOrEqual, have.tpl)
		assert.Equal(t, "must be greater or equal to 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.Zero(t, have.flags)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, msgGreaterOrEqual, have.tpl)
		assert.Equal(t, "must be greater or equal to 42", have.msg)
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("error - error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrCode, true)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `range-rule: spec argument "err_code" must be string, got bool`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, msgGreaterOrEqual, have.tpl)
		assert.Equal(t, "must be greater or equal to 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.Zero(t, have.flags)
	})

	t.Run("error - custom cmp is function invalid", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(spec.ArgSrc, true)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: spec argument \"src_go\" must be " +
			"verax.CompareFunc, got bool"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("custom cmp function", func(t *testing.T) {
		// --- Given ---
		fn := CompareFunc(func(any, any) (int, error) { return 0, nil })
		spc := spec.NewSpec(RangeRuleName).
			SetArg(ArgMode, "min").
			SetArg(spec.ArgValue, 42).
			SetArg(spec.ArgSrc, fn)

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, msgGreaterOrEqual, have.tpl)
		assert.Equal(t, "must be greater or equal to 42", have.msg)
		assert.Equal(t, ECInvRange, have.code)
		assert.Same(t, fn, have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})
}

func Test_rangeOutcome_tabular(t *testing.T) {
	tt := []struct {
		testN string

		mode   string
		result int
		want   bool
	}{
		{
			"a range must be greater than a value - value is less",
			"min-exclusive",
			-1,
			true,
		},
		{
			"a range must be greater than a value - value equal",
			"min-exclusive",
			0,
			false,
		},
		{
			"a range must be greater than a value - value greater",
			"min-exclusive",
			1,
			false,
		},

		// ---

		{
			"a range must be greater or equal than a value - value is less",
			"min",
			-1,
			true,
		},
		{
			"a range must be greater or equal than a value - value equal",
			"min",
			0,
			true,
		},
		{
			"a range must be greater or equal than a value - value greater",
			"min",
			1,
			false,
		},

		// ---

		{
			"a range must be less than a value - value is less",
			"max-exclusive",
			-1,
			false,
		},
		{
			"a range must be less than a value - value equal",
			"max-exclusive",
			0,
			false,
		},
		{
			"a range must be less than a value - value greater",
			"max-exclusive",
			1,
			true,
		},

		// ---

		{
			"a range must be less or equal than a value - value is less",
			"max",
			-1,
			false,
		},
		{
			"a range must be less or equal than a value - value equal",
			"max",
			0,
			true,
		},
		{
			"a range must be less or equal than a value - value greater",
			"max",
			1,
			true,
		},

		// ---

		{
			"unsupported operator",
			"unsupported",
			1,
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := rangeOutcome(tc.mode, tc.result)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_compareInt(t *testing.T) {
	t.Run("error - want is not integer", func(t *testing.T) {
		// --- When ---
		have, err := compareInt(1i+2, 1)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert complex128 to int64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not integer", func(t *testing.T) {
		// --- When ---
		have, err := compareInt(1, 1i+2)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert complex128 to int64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareInt_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"int - w is less than h", 1, 2, -1},
		{"int - w is equal to h", 1, 1, 0},
		{"int - w is greater than h", 1, 0, 1},

		{"int8 - w is less than h", int8(1), int8(2), -1},
		{"int8 - w is equal to h", int8(1), int8(1), 0},
		{"int8 - w is greater than h", int8(1), int8(0), 1},

		{"int16 - w is less than h", int16(1), int16(2), -1},
		{"int16 - w is equal to h", int16(1), int16(1), 0},
		{"int16 - w is greater than h", int16(1), int16(0), 1},

		{"int32 - w is less than h", int32(1), int32(2), -1},
		{"int32 - w is equal to h", int32(1), int32(1), 0},
		{"int32 - w is greater than h", int32(1), int32(0), 1},

		{"int64 - w is less than h", int64(1), int64(2), -1},
		{"int64 - w is equal to h", int64(1), int64(1), 0},
		{"int64 - w is greater than h", int64(1), int64(0), 1},

		{"duration - w is less than h", time.Second, time.Hour, -1},
		{"duration - w is equal to h", time.Second, time.Second, 0},
		{"duration - w is greater than h", time.Hour, time.Second, 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareInt(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareUInt(t *testing.T) {
	t.Run("error - want is not an unsigned integer", func(t *testing.T) {
		// --- When ---
		have, err := compareUInt(4.2, uint(1))

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert float64 to uint64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not an unsigned integer", func(t *testing.T) {
		// --- When ---
		have, err := compareUInt(uint(1), 4.2)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert float64 to uint64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareUint_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"uint - w is less than h", uint(1), uint(2), -1},
		{"uint - w is equal to h", uint(1), uint(1), 0},
		{"uint - w is greater than h", uint(1), uint(0), 1},

		{"uint8 - w is less than h", uint8(1), uint8(2), -1},
		{"uint8 - w is equal to h", uint8(1), uint8(1), 0},
		{"uint8 - w is greater than h", uint8(1), uint8(0), 1},

		{"uint16 - w is less than h", uint16(1), uint16(2), -1},
		{"uint16 - w is equal to h", uint16(1), uint16(1), 0},
		{"uint16 - w is greater than h", uint16(1), uint16(0), 1},

		{"uint32 - w is less than h", uint32(1), uint32(2), -1},
		{"uint32 - w is equal to h", uint32(1), uint32(1), 0},
		{"uint32 - w is greater than h", uint32(1), uint32(0), 1},

		{"uint64 - w is less than h", uint64(1), uint64(2), -1},
		{"uint64 - w is equal to h", uint64(1), uint64(1), 0},
		{"uint64 - w is greater than h", uint64(1), uint64(0), 1},

		{"uintptr - w is less than h", uintptr(1), uintptr(2), -1},
		{"uintptr - w is equal to h", uintptr(1), uintptr(1), 0},
		{"uintptr - w is greater than h", uintptr(1), uintptr(0), 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareUInt(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareFloat(t *testing.T) {
	t.Run("error - want is not float", func(t *testing.T) {
		// --- When ---
		have, err := compareFloat(1, 1i+2)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert complex128 to float64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not a float", func(t *testing.T) {
		// --- When ---
		have, err := compareFloat(1i+2, 1)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert complex128 to float64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareFloat_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"float32 - w is less than h", float32(1), float32(2), -1},
		{"float32 - w is equal to h", float32(1), float32(1), 0},
		{"float32 - w is greater than h", float32(1), float32(0), 1},

		{"float64 - w is less than h", float64(1), float64(2), -1},
		{"float64 - w is equal to h", float64(1), float64(1), 0},
		{"float64 - w is greater than h", float64(1), float64(0), 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareFloat(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareTime(t *testing.T) {
	t.Run("error - want is not time", func(t *testing.T) {
		// --- When ---
		have, err := compareTime(1, time.Now())

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert int to time.Time"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not time", func(t *testing.T) {
		// --- When ---
		have, err := compareTime(time.Now(), 1)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: cannot convert int to time.Time"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareTime_tabular(t *testing.T) {
	tim0 := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	tim1 := time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC)

	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"w is less than h", tim0, tim1, -1},
		{"w is equal to h", tim0, tim0, 0},
		{"w is greater than h", tim1, tim0, 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareTime(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareFor(t *testing.T) {
	t.Run("error - unsupported type", func(t *testing.T) {
		// --- When ---
		have, err := compareFor(test.Type{})

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "range-rule: unsupported type comparison: test.Type"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Nil(t, have)
	})
}

func Test_compareFor_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  any
		want CompareFunc
	}{
		{"int", 1, compareInt},
		{"int8", int8(1), compareInt},
		{"int16", int16(1), compareInt},
		{"int32", int32(1), compareInt},
		{"int64", int64(1), compareInt},

		{"uint", uint(1), compareUInt},
		{"uint8", uint8(1), compareUInt},
		{"uint16", uint16(1), compareUInt},
		{"uint32", uint32(1), compareUInt},
		{"uint64", uint64(1), compareUInt},
		{"uintptr", uintptr(1), compareUInt},

		{"float32", float32(1), compareFloat},
		{"float64", 1.0, compareFloat},

		{"time", time.Now(), compareTime},
		{"duration", time.Second, compareInt},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareFor(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Same(t, tc.want, have)
		})
	}
}

func Test_RangeRule_Spec_RangeRuleFromSpec_round_trip(t *testing.T) {
	t.Run("Min - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Min(42).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("Max - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Max(42).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("Min - With", func(t *testing.T) {
		// --- Given ---
		fn := CompareFunc(func(want, have any) (int, error) { return 0, nil })
		want := Min(42).With(fn)
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := RangeRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
