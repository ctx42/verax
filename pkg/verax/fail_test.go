// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Fail(t *testing.T) {
	t.Run("with error code", func(t *testing.T) {
		// --- When ---
		have := Fail("test err", "ECTst")

		// --- Then ---
		assert.True(t, have.condition)
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, "ECTst", have.code)
	})

	t.Run("with empty error code", func(t *testing.T) {
		// --- When ---
		have := Fail("test err", "")

		// --- Then ---
		assert.True(t, have.condition)
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, xrr.ECGeneric, have.code)
	})
}

func Test_FailRule_Validate(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := Fail("test err", "ECTst").When(false)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("returns the message and code", func(t *testing.T) {
		// --- Given ---
		r := Fail("test err", "ECTst")

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})
}

func Test_FailRule_When(t *testing.T) {
	// --- Given ---
	r := FailRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_FailRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := FailRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := FailRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
	})
}

func Test_FailRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := FailRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := AbsentRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_FailRule_Spec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := Fail("test err", "ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, FailRuleName, have.Name)
		wArgs := map[string]any{
			ArgErrMsg:  "test err",
			ArgErrCode: "ECTst",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("error - the error message cannot be empty", func(t *testing.T) {
		// --- Given ---
		r := FailRule{}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "fail-rule: error cannot have an empty message"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInternal, err)
		assert.Nil(t, have)
	})

	t.Run("empty code when code is ECGeneric", func(t *testing.T) {
		// --- Given ---
		r := Fail("test err", xrr.ECGeneric)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, FailRuleName, have.Name)
		wArgs := map[string]any{ArgErrMsg: "test err"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("empty code when code is empty", func(t *testing.T) {
		// --- Given ---
		r := FailRule{msg: "test err", code: ""}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, FailRuleName, have.Name)
		wArgs := map[string]any{ArgErrMsg: "test err"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("Fail - JSON representation", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()

		// --- When ---
		spc, err := Fail("test error message", "ECTst").Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(reg.EncodeSpec(spc))
		want := `{
			"name": "fail-rule",
			"args": {
				"err_code": {"type": "string", "value": "ECTst"},
				"err_msg": {"type": "string", "value": "test error message"}
			}
		}`
		assert.JSON(t, want, data)
	})
}

func Test_FailRuleFromSpec(t *testing.T) {
	t.Run("error - not error rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := FailRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `fail-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - message argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(FailRuleName)

		// --- When ---
		have, err := FailRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `fail-rule: spec missing required argument: err_msg`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("Fail", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(FailRuleName).
			SetArg(ArgErrMsg, "test err").
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := FailRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Fail("test err", "ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("with empty error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(FailRuleName).SetArg(ArgErrMsg, "test err")

		// --- When ---
		have, err := FailRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Fail("test err", xrr.ECGeneric)
		assert.Equal(t, wRule, have)
	})

	t.Run("error - error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(FailRuleName).
			SetArg(ArgErrMsg, "test err").
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := FailRuleFromSpec(spc)

		// --- Then ---
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		wMsg := `fail-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Empty(t, have)
	})
}

func Test_FailRule_Spec_FailRuleFromSpec_round_trip(t *testing.T) {
	// --- Given ---
	want := Fail("test msg", "ECTst")
	spc := must.Value(want.Spec())

	// --- When ---
	have, err := FailRuleFromSpec(spc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, want, have)
}
