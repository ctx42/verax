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

func Test_By(t *testing.T) {
	// --- Given ---
	fn := func(v any) error { return ErrTst }

	// --- When ---
	have := By(fn)

	// --- Then ---
	want := ByRule{
		fn:        fn,
		condition: true,
		msg:       "",
		code:      "",
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_ByRule_Validate(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		fn := func(any) error { return ErrTst }
		r := By(fn).When(false)

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil is ok", func(t *testing.T) {
		// --- Given ---
		r := By(nil)

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty is ok", func(t *testing.T) {
		// --- Given ---
		r := By(nil)

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("passes the argument to function", func(t *testing.T) {
		// --- Given ---
		var have any
		fn := func(h any) error { have = h; return nil }
		r := By(fn)

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "abc", have)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- Given ---
		fn := func(any) error { return ErrTst }
		r := By(fn)

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		fn := func(any) error { return NewError("fn msg", "ECFn") }
		r := By(fn).Message("test err")

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECFn", err)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		fn := func(any) error { return NewError("fn msg", "ECFn") }
		r := By(fn).Code("ECTst")

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "fn msg", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("custom error message and code", func(t *testing.T) {
		// --- Given ---
		fn := func(any) error { return NewError("fn msg", "ECFn") }
		r := By(fn).Message("test err").Code("ECTst")

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})
}

func Test_ByRule_When(t *testing.T) {
	// --- Given ---
	r := ByRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_ByRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := ByRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := ByRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_ByRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := ByRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := ByRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_ByRule_Spec(t *testing.T) {
	t.Run("By", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		r := By(fn)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, ByRuleName, have.Name)
		wArgs := map[string]any{
			spec.ArgSrc: RuleFunc(fn),
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		r := By(fn).Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, ByRuleName, have.Name)
		wArgs := map[string]any{
			ArgErrCode:  "ECTst",
			spec.ArgSrc: RuleFunc(fn),
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		r := By(fn).Message("test err")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, ByRuleName, have.Name)
		wArgs := map[string]any{
			ArgErrMsg:   "test err",
			spec.ArgSrc: RuleFunc(fn),
		}
		assert.Equal(t, wArgs, have.Args)
	})
}

func Test_ByRuleFromSpec(t *testing.T) {
	t.Run("error - not by rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `by-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - function argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(ByRuleName)

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `by-rule: spec missing required argument: src_go`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - custom function argument not RuleFunc", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(ByRuleName).
			SetArg(ArgMode, "bad-mode").
			SetArg(spec.ArgSrc, 42)

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `by-rule: spec argument "src_go" must be verax.RuleFunc, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		spc := spec.NewSpec(ByRuleName).
			SetArg(spec.ArgSrc, RuleFunc(fn)).
			SetArg(ArgErrMsg, "test err")

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := By(fn).Message("test err")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		spc := spec.NewSpec(ByRuleName).
			SetArg(spec.ArgSrc, RuleFunc(fn)).
			SetArg(ArgErrMsg, 42)

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `by-rule: spec argument "err_msg" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		spc := spec.NewSpec(ByRuleName).
			SetArg(spec.ArgSrc, RuleFunc(fn)).
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := By(fn)
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		spc := spec.NewSpec(ByRuleName).
			SetArg(spec.ArgSrc, RuleFunc(fn)).
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := By(fn).Code("ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error code not string", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		spc := spec.NewSpec(ByRuleName).
			SetArg(spec.ArgSrc, RuleFunc(fn)).
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `by-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return nil }
		spc := spec.NewSpec(ByRuleName).
			SetArg(spec.ArgSrc, RuleFunc(fn)).
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := ByRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := By(fn)
		assert.Equal(t, wRule, have)
	})
}

func Test_ByRule_Spec_ByRuleFromSpec_round_trip(t *testing.T) {
	// --- Given ---
	fn := func(v any) error { return nil }
	want := By(fn).Message("test msg").Code("ECTst")
	spc := must.Value(want.Spec())

	// --- When ---
	have, err := ByRuleFromSpec(spc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, want, have)
}
