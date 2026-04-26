// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Equal(t *testing.T) {
	t.Run("setup", func(t *testing.T) {
		// --- When ---
		have := Equal(42)

		// --- Then ---
		want := EqualRule{
			mode:      "equal",
			want:      42,
			condition: true,
			fn:        equal,
			tpl:       msgEqual,
			msg:       "must be equal to '42'",
			code:      ECNotEqual,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - not supported value", func(t *testing.T) {
		// --- When ---
		have := Equal(func() {})

		// --- Then ---
		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "equal-rule(equal): template render error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
	})
}

func Test_NotEqual(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		have := NotEqual(42)

		// --- Then ---
		want := EqualRule{
			mode:      "not-equal",
			want:      42,
			condition: true,
			fn:        notEqual,
			tpl:       msgNotEqual,
			msg:       "must not be equal to '42'",
			code:      ECEqual,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - not supported value", func(t *testing.T) {
		// --- When ---
		have := NotEqual(func() {})

		// --- Then ---
		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "equal-rule(not-equal): template render error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
	})
}

func Test_EqualField(t *testing.T) {
	// --- When ---
	have := EqualField(42, "field")

	// --- Then ---
	want := EqualRule{
		mode:      "equal",
		want:      42,
		condition: true,
		fn:        equal,
		tpl:       "must be equal to 'field'",
		msg:       "must be equal to 'field'",
		code:      ECNotEqual,
		sticky:    nil,
		flags:     flgCustomMsg,
	}
	assert.Equal(t, want, have)
}

func Test_NotEqualField(t *testing.T) {
	// --- When ---
	have := NotEqualField(42, "field")

	// --- Then ---
	want := EqualRule{
		mode:      "not-equal",
		want:      42,
		condition: true,
		fn:        notEqual,
		tpl:       "must not be equal to 'field'",
		msg:       "must not be equal to 'field'",
		code:      ECEqual,
		sticky:    nil,
		flags:     flgCustomMsg,
	}
	assert.Equal(t, want, have)
}

func Test_equal_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want any
		have any
	}{
		{"nil", nil, nil},
		{"int", 1, 1},
		{"string", "abc", "abc"},
		{"empty string", "", ""},
		{"float", 1.42, 1.42},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := equal(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_equal_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want any
		have any
	}{
		{"int", 1, 2},
		{"string", "abc", "xyz"},
		{"float", 1.42, 1.44},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := equal(tc.want, tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, "equal error", err)
			xrrtest.AssertCode(t, ECNotEqual, err)
		})
	}
}

func Test_notEqual_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want any
		have any
	}{
		{"nil", nil, 1},
		{"int", 1, 2},
		{"string", "abc", "xyz"},
		{"float", 1.42, 1.44},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := notEqual(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_notEqual_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want any
		have any
	}{
		{"nil", nil, nil},
		{"int", 1, 1},
		{"string", "abc", "abc"},
		{"empty string", "", ""},
		{"float", 1.42, 1.42},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := notEqual(tc.want, tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, "not equal error", err)
			xrrtest.AssertCode(t, ECEqual, err)
		})
	}
}

func Test_EqualRule_Validate(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := EqualRule{sticky: ErrTst}

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).When(false)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil ok", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty ok", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - with custom error message", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).Message("custom msg")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		xrrtest.AssertEqual(t, "custom msg (ECNotEqual)", err)
	})

	t.Run("error - with custom error code", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).Code("ECTst")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		xrrtest.AssertEqual(t, "must be equal to '42' (ECTst)", err)
	})

	t.Run("error - with custom error message and code", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).Message("custom msg").Code("ECTst")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		xrrtest.AssertEqual(t, "custom msg (ECTst)", err)
	})

	t.Run("arguments passed to EqualFunc", func(t *testing.T) {
		// --- Given ---
		var want, have any
		fn := func(w, h any) error { want = w; have = h; return nil }
		r := Equal(42).With(fn)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 42, want)
		assert.Equal(t, 44, have)
	})

	t.Run("error - error from EqualFunc", func(t *testing.T) {
		// --- Given ---
		fn := func(x, y any) error { return ErrTst }
		r := Equal(42).With(fn)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("error - custom error message with EqualFunc", func(t *testing.T) {
		// --- Given ---
		e := NewError("test msg", "ECTst")
		fn := func(x, y any) error { return e }
		r := Equal(42).With(fn).Message("custom msg")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.False(t, errors.Is(err, e))
		assert.ErrorEqual(t, "custom msg", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - custom error code with EqualFunc", func(t *testing.T) {
		// --- Given ---
		e := NewError("test msg", "ECTst")
		fn := func(x, y any) error { return e }
		r := Equal(42).With(fn).Code("ECCustom")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorIs(t, e, err)
		assert.ErrorEqual(t, "test msg", err)
		xrrtest.AssertCode(t, "ECCustom", err)
	})

	t.Run("error - custom error message and code with EqualFunc", func(t *testing.T) {
		// --- Given ---
		e := NewError("test msg", "ECTst")
		fn := func(x, y any) error { return e }
		r := Equal(42).With(fn).Message("custom msg").Code("ECTst")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.False(t, errors.Is(err, e))
		assert.ErrorEqual(t, "custom msg", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - NotEqual", func(t *testing.T) {
		// --- Given ---
		r := NotEqual(42)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		xrrtest.AssertEqual(t, "must not be equal to '42' (ECEqual)", err)
	})

	t.Run("successful validation", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_EqualRule_When(t *testing.T) {
	// --- Given ---
	r := EqualRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_EqualRule_With(t *testing.T) {
	t.Run("changes nothing when a sticky error is set", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }
		r := EqualRule{sticky: ErrTst}

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Same(t, ErrTst, have.sticky)
		assert.Zero(t, have.flags)
	})

	t.Run("equal to equal-by", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }
		r := Equal(42)

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Equal(t, "equal-by", have.mode)
		assert.Same(t, fn, have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})

	t.Run("equal-by to equal-by", func(t *testing.T) {
		// --- Given ---
		fn0 := func(any, any) error { return nil }
		fn1 := func(any, any) error { return nil }
		r := Equal(42).With(fn0)

		// --- When ---
		have := r.With(fn1)

		// --- Then ---
		assert.Equal(t, "equal-by", have.mode)
		assert.Same(t, fn1, have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})

	t.Run("not-equal to not-equal-by", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }
		r := NotEqual(42)

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Equal(t, "not-equal-by", have.mode)
		assert.Same(t, fn, have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})

	t.Run("not-equal-by to not-equal-by", func(t *testing.T) {
		// --- Given ---
		fn0 := func(any, any) error { return nil }
		fn1 := func(any, any) error { return nil }
		r := NotEqual(42).With(fn0)

		// --- When ---
		have := r.With(fn1)

		// --- Then ---
		assert.Equal(t, "not-equal-by", have.mode)
		assert.Same(t, fn1, have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})

	t.Run("success", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }
		r := EqualRule{mode: "equal"}

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Same(t, EqualFunc(fn), have.fn)
		assert.Equal(t, flgCustomFn, have.flags)
	})

	t.Run("error - unknown mode", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }
		r := EqualRule{mode: "unknown"}

		// --- When ---
		have := r.With(fn)

		// --- Then ---
		assert.Equal(t, "unknown", have.mode)
		assert.Nil(t, have.fn)
		assert.Zero(t, have.flags)

		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := `equal-rule: invalid rule mode: "unknown"`
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInvRuleMode, have.sticky)
	})
}

func Test_EqualRule_Message(t *testing.T) {
	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, msgEqual, have.tpl)
		assert.Equal(t, "must be equal to '42'", have.msg)
		assert.Zero(t, have.flags)
	})

	t.Run("when the sticky error is not nil", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)
		r.sticky = ErrTst

		// --- When ---
		have := r.Message("custom tpl {{.value}}")

		// --- Then ---
		assert.Equal(t, msgEqual, have.tpl)
		assert.Equal(t, "must be equal to '42'", have.msg)
		assert.Zero(t, have.flags)
	})

	t.Run("with value placeholder", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		have := r.Message("custom tpl {{.value}}")

		// --- Then ---
		assert.Equal(t, "custom tpl {{.value}}", have.tpl)
		assert.Equal(t, "custom tpl 42", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("without a value placeholder", func(t *testing.T) {
		// --- Given ---
		r := EqualRule{}

		// --- When ---
		have := r.Message("custom tpl")

		// --- Then ---
		assert.Equal(t, "custom tpl", have.tpl)
		assert.Equal(t, "custom tpl", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("error - template parse error", func(t *testing.T) {
		// --- Given ---
		r := EqualRule{mode: "mode"}

		// --- When ---
		have := r.Message("custom {{.")

		// --- Then ---
		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "equal-rule(mode): custom template parse error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
		assert.Zero(t, have.flags)
	})

	t.Run("error - with not supported custom placeholder", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		have := r.Message("custom tpl {{.custom}}")

		// --- Then ---
		assert.Equal(t, msgEqual, have.tpl)
		assert.Equal(t, "must be equal to '42'", have.msg)
		assert.Zero(t, have.flags)

		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "equal-rule(equal): custom template render error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
		assert.Zero(t, have.flags)
	})
}

func Test_EqualRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := EqualRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := EqualRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_EqualRule_Spec(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := EqualRule{sticky: ErrTst}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.Same(t, ErrTst, err)
		assert.Nil(t, have)
	})

	t.Run("mode equal", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EqualRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "equal", spec.ArgValue: 42}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode not-equal", func(t *testing.T) {
		// --- Given ---
		r := NotEqual(42)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EqualRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "not-equal", spec.ArgValue: 42}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode equal with EqualFunc", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }
		r := Equal(42).With(fn)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EqualRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:       "equal-by",
			spec.ArgValue: 42,
			spec.ArgSrc:   EqualFunc(fn),
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode not-equal with EqualFunc", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }
		r := NotEqual(42).With(fn)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EqualRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:       "not-equal-by",
			spec.ArgValue: 42,
			spec.ArgSrc:   EqualFunc(fn),
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		r := NotEqual(42).Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EqualRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:       "not-equal",
			spec.ArgValue: 42,
			ArgErrCode:    "ECTst",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).Message("test err")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, EqualRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:       "equal",
			spec.ArgValue: 42,
			ArgErrMsg:     "test err",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		r := EqualRule{mode: "unknown"}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `equal-rule: invalid rule mode: "unknown"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvRuleMode, err)
		assert.Nil(t, have)
	})
}

func Test_EqualRuleFromSpec(t *testing.T) {
	t.Run("error - not equal rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `equal-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - mode argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "equal-rule: spec missing required argument: mode"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - want argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).SetArg(ArgMode, "equal")

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "equal-rule: spec missing required argument: value"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "bad-mode").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `equal-rule: invalid spec rule mode: "bad-mode"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("mode equal", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Equal(42)
		assert.Equal(t, wRule, have)
	})

	t.Run("mode not-equal", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "not-equal").
			SetArg(spec.ArgValue, 42)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := NotEqual(42)
		assert.Equal(t, wRule, have)
	})

	t.Run("error - mode equal-by invalid EqualFunc", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal-by").
			SetArg(spec.ArgValue, 42).
			SetArg(spec.ArgSrc, 42)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `equal-rule: spec argument "src_go" must be verax.EqualFunc, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("mode equal with EqualFunc", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }

		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal-by").
			SetArg(spec.ArgValue, 42).
			SetArg(spec.ArgSrc, EqualFunc(fn))

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Equal(42).With(fn)
		assert.Equal(t, wRule, have)
	})

	t.Run("error - mode not-equal-by invalid EqualFunc", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "not-equal-by").
			SetArg(spec.ArgValue, 42).
			SetArg(spec.ArgSrc, 42)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `equal-rule: spec argument "src_go" must be verax.EqualFunc, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("mode not-equal with EqualFunc", func(t *testing.T) {
		// --- Given ---
		fn := func(any, any) error { return nil }

		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "not-equal-by").
			SetArg(spec.ArgValue, 42).
			SetArg(spec.ArgSrc, EqualFunc(fn))

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := NotEqual(42).With(fn)
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrMsg, "test err")

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Equal(42).Message("test err")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrMsg, 42)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `equal-rule: spec argument "err_msg" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Equal(42)
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Equal(42).Code("ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `equal-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(EqualRuleName).
			SetArg(ArgMode, "equal").
			SetArg(spec.ArgValue, 42).
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Equal(42)
		assert.Equal(t, wRule, have)
	})
}

func Test_EqualRule_Spec_EqualRuleFromSpec_round_trip(t *testing.T) {
	t.Run("Equal - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Equal(42).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("NotEqual - with message and code", func(t *testing.T) {
		// --- Given ---
		want := NotEqual(42).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("EqualField - with message and code", func(t *testing.T) {
		// --- Given ---
		want := EqualField(42, "field").Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("Equal with EqualFunc - with message and code", func(t *testing.T) {
		// --- Given ---
		fn := EqualFunc(func(any, any) error { return nil })
		want := Equal(42).With(fn).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("NotEqual with EqualFunc - with message and code", func(t *testing.T) {
		// --- Given ---
		fn := EqualFunc(func(any, any) error { return nil })
		want := NotEqual(42).With(fn).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := EqualRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
