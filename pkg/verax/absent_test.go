// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Nil(t *testing.T) {
	// --- Given ---
	have := Nil

	// --- Then ---
	want := AbsentRule{
		mode:      "nil",
		condition: true,
		msg:       msgReqNil,
		code:      ECReqNil,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_Empty(t *testing.T) {
	// --- Given ---
	have := Empty

	// --- Then ---
	want := AbsentRule{
		mode:      "empty",
		condition: true,
		msg:       msgReqEmpty,
		code:      ECReqEmpty,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_AbsentRule_Validate(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := AbsentRule{condition: false}

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_AbsentRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		mode string
		val  any
	}{
		{"nil", "nil", nil},
		{"nil pointer to string", "nil", pStringNil},
		{"nil pointer to int", "nil", pIntNil},
		{"nil pointer to time", "nil", pTimeNil},
		{"nil pointer to empty struct", "nil", pStructEmptyNil},
		{"nil declared slice", "nil", dSlice},
		{"nil declared map", "nil", dMap},

		{"nil", "empty", nil},
		{"empty string", "empty", ""},
		{"empty string instance", "empty", iStringEmpty},
		{"nil pointer to string", "empty", pStringNil},
		{"declared int", "empty", dInt},
		{"zero value time", "empty", iTimeZero},
		{"empty struct", "empty", iStructEmpty},
		{"declared empty struct", "empty", dStructEmpty},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := AbsentRule{condition: true, mode: tc.mode}

			// --- When ---
			err := r.Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_AbsentRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		mode string
		have any
	}{
		{"int", "nil", 123},
		{"string instance", "nil", iString},
		{"empty string", "nil", ""},
		{"declared string", "nil", dString},
		{"time", "nil", time.Now()},
		{"time", "nil", iTime},
		{"declared time", "nil", dTime},
		{"zero value time", "nil", time.Time{}},

		{"string", "empty", "abc"},
		{"string instance", "empty", iString},
		{"pointer to string instance", "empty", &iString},
		{"int", "empty", 123},
		{"time instance", "empty", time.Now()},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := AbsentRule{
				condition: true,
				mode:      tc.mode,
				msg:       "msg",
				code:      "code",
			}

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, "msg", err)
			xrrtest.AssertCode(t, "code", err)
		})
	}
}

func Test_AbsentRule_When(t *testing.T) {
	// --- Given ---
	r := AbsentRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_AbsentRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := AbsentRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := AbsentRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_AbsentRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := AbsentRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
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

func Test_AbsentRule_Spec(t *testing.T) {
	t.Run("mode nil", func(t *testing.T) {
		// --- Given ---
		r := Nil

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, AbsentRuleName, have.Name)
		assert.Equal(t, map[string]any{"mode": "nil"}, have.Args)
	})

	t.Run("mode empty", func(t *testing.T) {
		// --- Given ---
		r := Empty

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, AbsentRuleName, have.Name)
		assert.Equal(t, map[string]any{ArgMode: "empty"}, have.Args)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		r := Nil.Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, AbsentRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "nil", ArgErrCode: "ECTst"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		r := Nil.Message("test err")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, AbsentRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "nil", ArgErrMsg: "test err"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		r := AbsentRule{mode: "bad-mode"}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `absent-rule: invalid rule mode: "bad-mode"`, err)
		xrrtest.AssertCode(t, ECInvRuleMode, err)
		assert.Nil(t, have)
	})

	t.Run("Nil - JSON encode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()

		// --- When ---
		spc, err := Nil.Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(reg.EncodeSpec(spc))
		want := `{
			"name": "absent-rule",
			"args": {"mode": "nil"}
		}`
		assert.JSON(t, want, data)
	})

	t.Run("Nil - JSON decode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		reg.RegisterBuilders(Builders())
		data := []byte(`{
			"name": "absent-rule",
			"args": {"mode": "nil"}
		}`)

		// --- When ---
		have, err := reg.DecodeAndBuild(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Nil, have)
	})
}

func Test_AbsentRuleFromSpec(t *testing.T) {
	t.Run("error - not absent rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `absent-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - mode argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName)

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "absent-rule: spec missing required argument: mode"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).SetArg(ArgMode, "bad-mode")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `absent-rule: invalid spec rule mode: "bad-mode"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("mode nil", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).SetArg(ArgMode, "nil")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Nil
		assert.Equal(t, wRule, have)
	})

	t.Run("mode empty", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).SetArg(ArgMode, "empty")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Empty
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).
			SetArg(ArgMode, "empty").
			SetArg(ArgErrMsg, "test err")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Empty.Message("test err")
		assert.Equal(t, wRule, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).
			SetArg(ArgMode, "empty").
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Empty
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).
			SetArg(ArgMode, "nil").
			SetArg(ArgErrMsg, 42)

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `absent-rule: spec argument "err_msg" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("with custom error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).
			SetArg(ArgMode, "empty").
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Empty.Code("ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).
			SetArg(ArgMode, "empty").
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Empty
		assert.Equal(t, wRule, have)
	})

	t.Run("error - error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(AbsentRuleName).
			SetArg(ArgMode, "nil").
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `absent-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})
}

func Test_AbsentRule_Spec_AbsentRuleFromSpec_round_trip(t *testing.T) {
	t.Run("Nil - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Nil.Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("Empty - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Nil.Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := AbsentRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
